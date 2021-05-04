package store

import (
	"github.com/gasparian/money-transfers-api/internal/app/models"
)

// Store ...
type Store interface {
	InsertAccount(*models.Account) error
	GetBalance(int64) (float64, error)
	Deposit(*models.Transfer) (float64, error)
	Withdraw(*models.Transfer) (float64, error)
	Transfer(*models.Transfer) (*models.TransferResult, error)
	DeleteAccount(int64) error
	GetTransfersHistory(*models.TransferHisotoryRequest) ([]models.Transfer, error)
}
