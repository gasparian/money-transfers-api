package sqlstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/gasparian/money-transfers-api/internal/app/models"
	_ "github.com/mattn/go-sqlite3"
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
		balance INTEGER CHECK(balance >= 0)
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
	    	timesamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    	from_account_id INTEGER NOT NULL,
	    	to_account_id INTEGER NOT NULL,
	    	amount INTEGER CHECK(amount > 0)
	    );`,
		`CREATE INDEX idx_from_account_id ON transfer(from_account_id)`,
		`CREATE INDEX idx_to_account_id ON transfer(to_account_id)`,
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

// DropTables removes table from the db
func (s *Store) DropTable(tableName string) error {
	_, err := s.db.Exec("DROP TABLE IF EXISTS ?", tableName)
	if err != nil {
		return err
	}
	return nil
}

// InsertAccount inserts new account into the accounts table and returns row's id
func (s *Store) InsertAccount(acc *models.Account) error {
	err := s.db.QueryRow(
		"INSERT INTO account(balance) VALUES ($1) ",
		acc.Balance,
	).Scan(&acc.AccountID)
	if err != nil {
		return err
	}
	return nil
}

// GetBalance returns account balance by it's id
func (s *Store) GetBalance(accountID int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	var balance int64
	err := s.db.QueryRowContext(
		ctx,
		"SELECT balance FROM account WHERE account_id=$1",
		accountID,
	).Scan(&balance)
	if err != nil {
		return balance, err
	}
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
	res := models.TransferResult{Transfer: *tr}
	err = tx.QueryRow(
		"UPDATE account SET balance = balance + $1 WHERE account_id=$2 RETURNING balance",
		-tr.Amount,
		tr.FromAccountID,
	).Scan(&res.FromAccountIDBalance)
	if err != nil {
		tx.Rollback()
		return nil, nil
	}
	err = tx.QueryRow(
		"UPDATE account SET balance = balance + $1 WHERE account_id=$2 RETURNING balance",
		tr.Amount,
		tr.ToAccountID,
	).Scan(&res.ToAccountIDBalance)
	if err != nil {
		tx.Rollback()
		return nil, nil
	}
	err = tx.QueryRow(
		"INSERT INTO transfer(from_account_id, to_account_id, amount) VALUES ($1, $2, $3) RETURNING (transfer_id, timestamp)",
		tr.FromAccountID,
		tr.ToAccountID,
		tr.Amount,
	).Scan(&res.Transfer.TransferID, &res.Transfer.Timestamp)
	if err != nil {
		tx.Rollback()
		return nil, nil
	}
	tx.Commit()
	return &res, nil
}

// DeleteAccount removes account from the accounts table
func (s *Store) DeleteAccount(accountID int64) error {
	_, err := s.db.Exec(
		"DELETE FROM account WHERE id=$1",
		accountID,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetTranscationsHistory retunrs array of transcation for certain period of time
func (s *Store) GetTranscationsHistory(accountID, nDays int64) ([]models.Transfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	row, err := s.db.QueryContext(
		ctx,
		"SELECT * FROM transfer WHERE timestamp < date('now', '-$1 days') AND (from_account_id=$2 OR to_account_id=$2)",
		nDays,
		accountID,
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
