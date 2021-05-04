package models

import (
	"time"
)

// Account holds id and amount of money
type Account struct {
	AccountID int64   `json:"account_id"`
	Balance   float64 `json:"balance"`
}

// Transfer holds data needed to perform money transfer
type Transfer struct {
	TransferID    int64     `json:"transfer_id"`
	Timestamp     time.Time `json:"timestamp"`
	FromAccountID int64     `json:"from_account_id"`
	ToAccountID   int64     `json:"to_account_id"`
	Amount        float64   `json:"amount"`
}

// TransferResult hodls data about performed transfer and the new balances
type TransferResult struct {
	FromAccountID        int64   `json:"from_account_id"`
	ToAccountID          int64   `json:"to_account_id"`
	FromAccountIDBalance float64 `json:"from_account_id_balance"`
	ToAccountIDBalance   float64 `json:"to_account_id_balance"`
	TransferID           int64   `json:"transfer_id"`
}

// TransferHisotoryRequest holds account id and the number of days to look at the past from now
type TransferHisotoryRequest struct {
	AccountID int64 `json:"account_id"`
	NDays     int64 `json:"n_days"`
}
