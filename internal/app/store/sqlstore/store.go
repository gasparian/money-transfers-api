package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gasparian/money-transfers-api/internal/app/models"
	_ "github.com/mattn/go-sqlite3"
)

var (
	accountsArrayEmptyErr = errors.New("Accounts array is empty")
)

// Store object holds db instance
type Store struct {
	db           *sql.DB
	queryTimeout time.Duration
}

func newDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// New creates new instance of the db and create needed tables
func New(dbPath string, queryTimeout uint32) (*Store, error) {
	db, err := newDB(dbPath)
	if err != nil {
		return nil, err
	}
	s := &Store{
		db:           db,
		queryTimeout: time.Duration(queryTimeout) * time.Second,
	}
	s.createAccountsTable()
	s.createTransfersTable()
	return s, nil
}

// Close closes underlying db connection
func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) createAccountsTable() error {
	q := `CREATE TABLE IF NOT EXISTS account (
		account_id INTEGER NOT NULL PRIMARY KEY,
		balance REAL CHECK(balance >= 0.0)
	);`
	_, err := s.db.Exec(q)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) createTransfersTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	queries := []string{
		`CREATE TABLE IF NOT EXISTS transfer (
	    	transfer_id INTEGER NOT NULL PRIMARY KEY,
	    	timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    	from_account_id INTEGER,
	    	to_account_id INTEGER,
	    	amount REAL CHECK(amount > 0.0)
	    );`,
		`CREATE INDEX IF NOT EXISTS idx_from_account_id ON transfer(from_account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_to_account_id ON transfer(to_account_id)`,
	}
	for _, q := range queries {
		_, err := tx.Exec(q)
		if err != nil {
			tx.Rollback()
			return nil
		}
	}
	tx.Commit()
	return nil
}

// dropTables removes table from the db
// non-exposed method, because of potential sql-injections
func (s *Store) dropTable(tableName string) error {
	_, err := s.db.Exec(
		fmt.Sprintf(
			"DROP TABLE IF EXISTS %s",
			tableName,
		),
	)
	if err != nil {
		return err
	}
	return nil
}

// InsertAccount inserts new account into the accounts table and returns row's id
func (s *Store) InsertAccount(acc *models.Account) error {
	res, err := s.db.Exec(
		"INSERT INTO account(balance) VALUES (?)",
		acc.Balance,
	)
	if err != nil {
		return err
	}
	acc.AccountID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

// GetBalance returns account balance by it's id
func (s *Store) GetBalance(accountID int64) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	var balance float64
	err := s.db.QueryRowContext(
		ctx,
		"SELECT balance FROM account WHERE account_id=?",
		accountID,
	).Scan(&balance)
	if err != nil {
		return balance, err
	}
	return balance, nil
}

func getAccountsTx(tx *sql.Tx, accounts []int64) (map[int64]float64, error) {
	if len(accounts) == 0 {
		return nil, accountsArrayEmptyErr
	}
	args := make([]interface{}, len(accounts))
	for i, acc := range accounts {
		args[i] = acc
	}
	query := "SELECT * FROM account WHERE account_id IN (?" + strings.Repeat(",?", len(args)-1) + ")"
	row, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	res := make(map[int64]float64)
	for row.Next() {
		var accID int64
		var balance float64
		row.Scan(
			&accID,
			&balance,
		)
		res[accID] = balance
	}
	return res, nil
}

// Deposit adds money to the account
func (s *Store) Deposit(tr *models.Transfer) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(
		"UPDATE account SET balance = balance + ? WHERE account_id=?",
		tr.Amount,
		tr.ToAccountID,
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	_, err = tx.Exec(
		"INSERT INTO transfer(from_account_id, to_account_id, amount) VALUES (0, ?, ?)",
		tr.ToAccountID,
		tr.Amount,
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	accs, err := getAccountsTx(tx, []int64{tr.ToAccountID})
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	balance := accs[tr.ToAccountID]
	tx.Commit()
	return balance, nil
}

// Withdraw pulls money from the account
func (s *Store) Withdraw(tr *models.Transfer) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(
		"UPDATE account SET balance = balance + ? WHERE account_id=?",
		-tr.Amount,
		tr.FromAccountID,
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	_, err = tx.Exec(
		"INSERT INTO transfer(from_account_id, to_account_id, amount) VALUES (?, 0, ?)",
		tr.FromAccountID,
		tr.Amount,
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	accs, err := getAccountsTx(tx, []int64{tr.FromAccountID})
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()
	balance := accs[tr.FromAccountID]
	return balance, nil
}

// Transfer transfers money from one account to another; writes transfer info into the transfers table
func (s *Store) Transfer(tr *models.Transfer) (*models.TransferResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	trRes := models.TransferResult{
		FromAccountID: tr.FromAccountID,
		ToAccountID:   tr.ToAccountID,
	}
	_, err = tx.Exec(
		"UPDATE account SET balance = balance + ? WHERE account_id=?",
		-tr.Amount,
		tr.FromAccountID,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	_, err = tx.Exec(
		"UPDATE account SET balance = balance + ? WHERE account_id=?",
		tr.Amount,
		tr.ToAccountID,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	res, err := tx.Exec(
		"INSERT INTO transfer(from_account_id, to_account_id, amount) VALUES (?, ?, ?)",
		tr.FromAccountID,
		tr.ToAccountID,
		tr.Amount,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	trRes.TransferID, err = res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	accs, err := getAccountsTx(tx, []int64{trRes.ToAccountID, trRes.FromAccountID})
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	trRes.ToAccountIDBalance = accs[trRes.ToAccountID]
	trRes.FromAccountIDBalance = accs[trRes.FromAccountID]
	return &trRes, nil
}

// DeleteAccount removes account from the accounts table
func (s *Store) DeleteAccount(accountID int64) error {
	_, err := s.db.Exec(
		"DELETE FROM account WHERE account_id=?",
		accountID,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetTransfersHistory retunrs array of transcations for the requested period of time
func (s *Store) GetTransfersHistory(req *models.TransferHisotoryRequest) ([]models.Transfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	row, err := s.db.QueryContext(
		ctx,
		fmt.Sprintf(`SELECT * FROM transfer WHERE 
		timestamp >= date('now', '-%v day') AND 
		(from_account_id=$1 OR to_account_id=$1)`, req.NDays),
		req.AccountID,
	)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var res []models.Transfer
	for row.Next() {
		tmpRecord := models.Transfer{}
		row.Scan(
			&tmpRecord.TransferID,
			&tmpRecord.Timestamp,
			&tmpRecord.FromAccountID,
			&tmpRecord.ToAccountID,
			&tmpRecord.Amount,
		)
		res = append(res, tmpRecord)
	}
	return res, nil
}
