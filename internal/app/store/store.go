package store

import (
	"github.com/gasparian/money-transfers-api/internal/app/model"
)

// Store ...
type Store interface {
	Create(*model.Account) error
	Get(*model.Account) error
	Update(*model.Account) error
	Delete(*model.Account) error
}
