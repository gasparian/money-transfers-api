package model

type Account struct {
	ID       uint64 `json:"id"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	Currency string `json:"currency"`
	Balance  uint64 `json:"balance"`
}
