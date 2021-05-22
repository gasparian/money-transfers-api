package models

import (
	"time"
)

// Account holds info about account that stored in the db
type Account struct {
	CreatedAt time.Time
	AccountID int64
	Balance   int64
}

// Transaction holds data needed to perform money transfer
type Transaction struct {
	TransactionID int64
	Timestamp     time.Time
	FromAccountID int64
	ToAccountID   int64
	Amount        int64
}
