package apiserver

import (
	"time"
)

// AccountJsonView holds id and amount of money
type AccountJsonView struct {
	AccountID int64   `json:"account_id"`
	Balance   float64 `json:"balance"`
}

// AccountIDJsonView ...
type AccountIDJsonView struct {
	ID int64 `json:"account_id"`
}

// TransactionJsonView holds data needed to perform money transfer
type TransactionJsonView struct {
	Timestamp     time.Time `json:"timestamp"`
	FromAccountID int64     `json:"from_account_id"`
	ToAccountID   int64     `json:"to_account_id"`
	Amount        float64   `json:"amount"`
}
