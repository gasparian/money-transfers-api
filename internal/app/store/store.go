package store

import (
	"github.com/gasparian/money-transfers-api/internal/app/models"
)

// Store ...
type Store interface {
	InsertAccount(models.MoneyAmount) (models.Account, error)
	DeleteAccount(accountId int64) error
	GetAccount(accountId int64) (models.Account, error)
	TransferMoney(accountToId, accountFromId int64, amount models.MoneyAmount) error
	GetTransactionsHistory(accountId, nLastDays, limit int64) ([]models.Transaction, error)
}
