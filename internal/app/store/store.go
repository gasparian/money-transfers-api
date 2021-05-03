package store

import (
	"github.com/gasparian/money-transfers-api/internal/app/models"
)

// Store ...
type Store interface {
	InsertAccount(*models.Account) error
	GetBalance(int64) (int64, error)
	Transfer(*models.Transfer) (*models.TransferResult, error)
	DeleteAccount(int64) error
	GetTranscationsHistory(int64, int64) ([]models.Transfer, error)
}
