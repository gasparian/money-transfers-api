package store

import (
	"github.com/gasparian/money-transfers-api/internal/app/models"
)

// Store ...
type Store interface {
	InsertAccount(balance int64) (models.Account, error)
	DeleteAccount(accountId int64) error
	GetAccount(accountId int64) (models.Account, error)
	TransferMoney(accountToId, accountFromId, amount int64) error
	GetTransactionsHistory(accountId, nLastDays, limit int64) ([]models.Transaction, error)
}
