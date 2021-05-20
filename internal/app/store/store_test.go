package store_test

import (
	"errors"
	"github.com/gasparian/money-transfers-api/internal/app/models"
	"github.com/gasparian/money-transfers-api/internal/app/store"
	"github.com/gasparian/money-transfers-api/internal/app/store/kvstore"
	"github.com/gasparian/money-transfers-api/internal/app/store/sqlstore"
	"os"
	"testing"
)

var (
	invalidBalanceValueErr  = errors.New("Invalid balance value")
	transactionCorruptedErr = errors.New("Transaction corrupted")
)

func testStore(store store.Store, t *testing.T) {
	t.Run("InsertAccount", func(t *testing.T) {
		var balance int64 = 1005
		acc, err := store.InsertAccount(balance)
		if err != nil {
			t.Fatal(err)
		}
		if balance != acc.Balance {
			t.Error(invalidBalanceValueErr)
		}
	})

	t.Run("DeleteAccount", func(t *testing.T) {
		var balance int64 = 1005
		acc, err := store.InsertAccount(balance)
		if err != nil {
			t.Fatal(err)
		}
		err = store.DeleteAccount(acc.AccountID)
		if err != nil {
			t.Error(err)
		}
		_, err = store.GetAccount(acc.AccountID)
		if err == nil {
			t.Error(invalidBalanceValueErr)
		}
	})

	t.Run("TransferGetAccount", func(t *testing.T) {
		accFrom, err := store.InsertAccount(10000)
		if err != nil {
			t.Fatal(err)
		}
		accTo, err := store.InsertAccount(1000)
		if err != nil {
			t.Fatal(err)
		}
		err = store.TransferMoney(
			accTo.AccountID,
			accFrom.AccountID,
			9000,
		)
		if err != nil {
			t.Fatal(err)
		}
		accToNew, err := store.GetAccount(accTo.AccountID)
		if err != nil {
			t.Fatal(err)
		}
		accFromNew, err := store.GetAccount(accFrom.AccountID)
		if err != nil {
			t.Fatal(err)
		}
		if !(accToNew.Balance == accFrom.Balance && accFromNew.Balance == accTo.Balance) {
			t.Error(transactionCorruptedErr)
		}
	})

	t.Run("TransferNegativeResult", func(t *testing.T) {
		accFrom, err := store.InsertAccount(50)
		if err != nil {
			t.Fatal(err)
		}
		accTo, err := store.InsertAccount(1000)
		if err != nil {
			t.Fatal(err)
		}
		err = store.TransferMoney(
			accTo.AccountID,
			accFrom.AccountID,
			1150,
		)
		if err == nil {
			t.Error(transactionCorruptedErr)
		}
	})

	t.Run("GetTransactions", func(t *testing.T) {
		accFrom, err := store.InsertAccount(10000)
		if err != nil {
			t.Fatal(err)
		}
		accTo, err := store.InsertAccount(0)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 5; i++ {
			err = store.TransferMoney(
				accTo.AccountID,
				accFrom.AccountID,
				2000,
			)
			if err != nil {
				t.Fatal(err)
			}
		}

		transactions, err := store.GetTransactionsHistory(
			accTo.AccountID, 1, 3,
		)
		if err != nil {
			t.Error(err)
		}
		if len(transactions) != 3 {
			t.Error(transactionCorruptedErr)
		}
		var summ int64 = 0
		for _, transaction := range transactions {
			summ += transaction.Amount
		}
		if summ != 6000 {
			t.Fatal(transactionCorruptedErr)
		}
		accToNew, err := store.GetAccount(accTo.AccountID)
		if err != nil {
			t.Fatal(err)
		}
		accFromNew, _ := store.GetAccount(accFrom.AccountID)
		if accFrom.Balance != accToNew.Balance || accTo.Balance != accFromNew.Balance {
			t.Error(transactionCorruptedErr)
		}
	})
}

func testStoreConcurrentTransfer(store store.Store, t *testing.T) {
	accFrom := models.Account{Balance: 10000}
	accFrom, err := store.InsertAccount(accFrom.Balance)
	if err != nil {
		t.Fatal(err)
	}
	accTo := models.Account{}
	accTo, err = store.InsertAccount(accTo.Balance)
	if err != nil {
		t.Fatal(err)
	}

	n := 100
	errs := make(chan error)
	for i := 0; i < n; i++ {
		go func(accToId, accFromId int64) {
			err := store.TransferMoney(
				accToId,
				accFromId,
				100,
			)
			errs <- err
		}(accTo.AccountID, accFrom.AccountID)
	}

	for i := 0; i < n; i++ {
		err := <-errs
		if err != nil {
			t.Fatal(err)
		}
	}

	accToNew, err := store.GetAccount(accTo.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	accFromNew, err := store.GetAccount(accFrom.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	if accToNew.Balance <= accFromNew.Balance {
		t.Fatal(invalidBalanceValueErr)
	}
}

func TestSqlStore(t *testing.T) {
	dbPath := "/tmp/tets.db"
	store, err := sqlstore.New(dbPath, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	defer os.RemoveAll(dbPath)

	testStore(store, t)
	testStoreConcurrentTransfer(store, t)
}

func TestKVStore(t *testing.T) {
	store := kvstore.New()
	testStore(store, t)
	testStoreConcurrentTransfer(store, t)
}
