package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	s.createTransactionsTable()
	return s, nil
}

// Close closes underlying db connection
func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) createAccountsTable() error {
	q := `CREATE TABLE IF NOT EXISTS account (
	    created_at TIMESTAMP DEFAULT(STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
		account_id INTEGER NOT NULL PRIMARY KEY,
		balance INTEGER,
        CHECK(balance >= 0)
	);`
	_, err := s.db.Exec(q)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) createTransactionsTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	queries := []string{
		`CREATE TABLE IF NOT EXISTS transactions (
	    	transaction_id INTEGER NOT NULL PRIMARY KEY,
	    	timestamp TIMESTAMP DEFAULT(STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	    	from_account_id INTEGER,
	    	to_account_id INTEGER,
	    	amount INTEGER,
            CHECK(amount >= 0)
	    );`,
		`CREATE INDEX IF NOT EXISTS idx_from_account_id ON transactions(from_account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_to_account_id ON transactions(to_account_id)`,
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

// InsertAccount inserts new account into the accounts table and returns Account model
func (s *Store) InsertAccount(balance int64) (models.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	var acc models.Account
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return acc, err
	}
	res, err := tx.Exec(
		"INSERT INTO account(balance) VALUES (?)",
		balance,
	)
	if err != nil {
		tx.Rollback()
		return acc, err
	}
	accId, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return acc, err
	}
	err = tx.QueryRowContext(
		ctx,
		"SELECT * FROM account WHERE account_id=?",
		accId,
	).Scan(
		&acc.CreatedAt,
		&acc.AccountID,
		&acc.Balance,
	)
	if err != nil {
		tx.Rollback()
		return acc, err
	}
	tx.Commit()
	return acc, nil
}

// DeleteAccount removes account from the accounts table
func (s *Store) DeleteAccount(accId int64) error {
	_, err := s.db.Exec(
		"DELETE FROM account WHERE account_id=?",
		accId,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetAccount returns account model
func (s *Store) GetAccount(accId int64) (models.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	var acc models.Account
	err := s.db.QueryRowContext(
		ctx,
		"SELECT * FROM account WHERE account_id=?",
		accId,
	).Scan(
		&acc.CreatedAt,
		&acc.AccountID,
		&acc.Balance,
	)
	return acc, err
}

// TransferMoney transfers money from one account to another; writes transfer info into the transfers table
func (s *Store) TransferMoney(accountToId, accountFromId, amount int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	updateBalanceQuery := "UPDATE account SET balance = balance + ? WHERE account_id=?"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		updateBalanceQuery,
		-amount,
		accountFromId,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(
		updateBalanceQuery,
		amount,
		accountToId,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(
		"INSERT INTO transactions(from_account_id, to_account_id, amount) VALUES (?, ?, ?)",
		accountFromId,
		accountToId,
		amount,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetTransactionsHistory retunrs array of transcations for the requested period of time
func (s *Store) GetTransactionsHistory(accountId, nLastdays, limit int64) ([]models.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
	defer cancel()

	row, err := s.db.QueryContext(
		ctx,
		fmt.Sprintf(`SELECT * FROM transactions WHERE 
		timestamp >= date('now', '-%v day') AND 
		(from_account_id=$1 OR to_account_id=$1) LIMIT $2`, nLastdays),
		accountId,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var res []models.Transaction
	for row.Next() {
		tmpRecord := models.Transaction{}
		row.Scan(
			&tmpRecord.TransactionID,
			&tmpRecord.Timestamp,
			&tmpRecord.FromAccountID,
			&tmpRecord.ToAccountID,
			&tmpRecord.Amount,
		)
		res = append(res, tmpRecord)
	}
	return res, nil
}
