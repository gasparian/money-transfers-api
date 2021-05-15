package kvstore

import (
	"errors"
	"github.com/gasparian/money-transfers-api/internal/app/models"
	"sort"
	"sync"
	"time"
)

var (
	accNotFoundErr        = errors.New("Account not found")
	notEnoghMoneyOnAccErr = errors.New("There is no enough money on account to complete a transaction")
)

type ConcurrentAccount struct {
	models.Account
	mx sync.RWMutex
}

type KVStore struct {
	mx               sync.RWMutex
	accIncID         int64
	transactionIncID int64
	accounts         map[int64]*ConcurrentAccount
	transactions     map[int64]models.Transaction
}

func New() *KVStore {
	return &KVStore{
		accounts:     make(map[int64]*ConcurrentAccount),
		transactions: make(map[int64]models.Transaction),
	}
}

func (s *KVStore) InsertAccount(balance models.MoneyAmount) (models.Account, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.accIncID++
	acc := &ConcurrentAccount{
		Account: models.Account{
			AccountID: s.accIncID,
			CreatedAt: time.Now(),
			Balance:   balance,
		},
	}
	s.accounts[s.accIncID] = acc
	return acc.Account, nil
}

func (s *KVStore) DeleteAccount(accId int64) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.accIncID--
	delete(s.accounts, accId)
	return nil
}

func (s *KVStore) GetAccount(accId int64) (models.Account, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	acc, ok := s.accounts[accId]
	if !ok {
		return models.Account{}, accNotFoundErr
	}
	return acc.Account, nil
}

func (s *KVStore) TransferMoney(accountToId, accountFromId int64, amount models.MoneyAmount) error {
	s.mx.RLock()
	accTo, ok := s.accounts[accountToId]
	if !ok {
		s.mx.RUnlock()
		return accNotFoundErr
	}
	accFrom, ok := s.accounts[accountFromId]
	if !ok {
		s.mx.RUnlock()
		return accNotFoundErr
	}
	s.mx.RUnlock()

	accFrom.mx.RLock()
	if models.CompareMoney(&accFrom.Balance, &amount) < 0 {
		accFrom.mx.RUnlock()
		return notEnoghMoneyOnAccErr
	}
	accFrom.mx.RUnlock()

	accs := []*ConcurrentAccount{accTo, accFrom}
	sort.Slice(accs, func(i, j int) bool {
		return accs[i].AccountID < accs[j].AccountID
	})

	for _, acc := range accs {
		acc.mx.Lock()
	}
	models.AddMoney(&accFrom.Balance, &accTo.Balance, &amount)
	for _, acc := range accs {
		acc.mx.Unlock()
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	s.transactionIncID++
	s.transactions[s.transactionIncID] = models.Transaction{
		TransactionID: s.transactionIncID,
		Timestamp:     time.Now(),
		FromAccountID: accFrom.AccountID,
		ToAccountID:   accTo.AccountID,
		Amount:        amount,
	}
	return nil
}

func (s *KVStore) GetTransactionsHistory(accountId, nLastdays, limit int64) ([]models.Transaction, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	limConv := int(limit)
	nLastdaysConv := int(nLastdays)
	tr := make([]models.Transaction, 0)
	// NOTE: here I used just run the full scan against all transactions
	for _, v := range s.transactions {
		accountIDMatches := v.ToAccountID == accountId || v.FromAccountID == accountId
		dateMatches := v.Timestamp.After(time.Now().AddDate(0, 0, -nLastdaysConv))
		if accountIDMatches && dateMatches {
			tr = append(tr, v)
		}
		if len(tr) >= limConv {
			break
		}
	}
	return tr, nil
}
