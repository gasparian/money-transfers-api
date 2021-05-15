package apiserver

import (
	"time"
)

// MoneyAmountJsonView holds id and amount of money
type MoneyAmountJsonView struct {
	Integer  int64 `json:"integer"`
	Fraction int64 `json:"fraction"`
}

// AccountJsonView holds id and amount of money
type AccountJsonView struct {
	AccountID int64               `json:"account_id"`
	Balance   MoneyAmountJsonView `json:"balance"`
}

// AccountIDJsonView ...
type AccountIDJsonView struct {
	ID int64 `json:"account_id"`
}

// TransactionJsonView holds data needed to perform money transfer
type TransactionJsonView struct {
	Timestamp     time.Time           `json:"timestamp"`
	FromAccountID int64               `json:"from_account_id"`
	ToAccountID   int64               `json:"to_account_id"`
	Amount        MoneyAmountJsonView `json:"amount"`
}
