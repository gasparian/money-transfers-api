package sqlstore

import (
	"database/sql"

	"github.com/gasparian/money-transfers-api/internal/app/model"
	"github.com/gasparian/money-transfers-api/internal/app/store"
)

// Store ...
type Store struct {
	db *sql.DB
}

// New ...
func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// Create ...
func (s *Store) Create(acc *model.Account) error {
	return s.db.QueryRow(
		"INSERT INTO users (email, currency, balance) VALUES ($1, $2, $3) RETURNING id",
		acc.Email,
		acc.Currency,
		acc.Balance,
	).Scan(&acc.ID)
}

// Get ...
func (s *Store) Get(acc *model.Account) error {
	return nil
}

// Update ...
func (s *Store) Update(acc *model.Account) error {
	return nil
}

// Delete ...
func (s *Store) Delete(acc *model.Account) error {
	return nil
}
