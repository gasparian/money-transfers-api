package models

import (
	"time"
)

// MoneyAmount sum of money represented in two parts
type MoneyAmount struct {
	Integer  int64
	Fraction int64
}

func AddMoney(from, to, amount *MoneyAmount) {
	from.Integer -= amount.Integer
	from.Fraction -= amount.Fraction
	to.Integer += amount.Integer
	to.Fraction += amount.Fraction
}

func CompareMoney(l, r *MoneyAmount) int {
	if l.Integer > r.Integer {
		return 1
	}
	if l.Integer < r.Integer {
		return -1
	}
	if l.Integer == r.Integer {
		if l.Fraction > r.Fraction {
			return 1
		}
		if l.Fraction < r.Fraction {
			return -1
		}
	}
	return 0
}

func SumMoney(l, r *MoneyAmount) *MoneyAmount {
	return &MoneyAmount{
		Integer:  l.Integer + r.Integer,
		Fraction: l.Fraction + r.Fraction,
	}
}

// Account holds info about account that stored in the db
type Account struct {
	CreatedAt time.Time
	AccountID int64
	Balance   MoneyAmount
}

// IsEqualAccounts ...
func IsEqualAccounts(l, r *Account) bool {
	accId := l.AccountID == r.AccountID
	balance := CompareMoney(&l.Balance, &r.Balance) == 0
	date := l.CreatedAt == r.CreatedAt
	return accId && balance && date
}

// Transaction holds data needed to perform money transfer
type Transaction struct {
	TransactionID int64
	Timestamp     time.Time
	FromAccountID int64
	ToAccountID   int64
	Amount        MoneyAmount
}
