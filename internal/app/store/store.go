package store

import (
	"github.com/gasparian/money-transfers-api/internal/app/models"
)

// Store ...
type Store interface {
	InsertAccount(*models.Account) error
	GetBalance(*models.Account) error
	Deposit(*models.Transfer) (*models.Account, error)
	Withdraw(*models.Transfer) (*models.Account, error)
	Transfer(*models.Transfer) (*models.TransferResult, error)
	DeleteAccount(*models.Account) error
	GetTransfersHistory(*models.TransferHisotoryRequest) ([]models.Transfer, error)
}
