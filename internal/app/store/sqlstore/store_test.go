package sqlstore

import (
	"errors"
	"github.com/gasparian/money-transfers-api/internal/app/models"
	"math"
	"os"
	"testing"
)

const (
	tol = 1e-4
)

var (
	invalidBalanceValueErr = errors.New("Invalid balance value")
	transferCorruptedErr   = errors.New("Transfer corrupted")
)

func TestStore(t *testing.T) {
	dbPath := "/tmp/tets.db"
	store, err := New(dbPath, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	defer os.RemoveAll(dbPath)

	t.Run("InsertAccount", func(t *testing.T) {
		var initBalance float64 = 10
		acc := &models.Account{Balance: initBalance}
		err := store.InsertAccount(acc)
		if err != nil {
			t.Fatal(err)
		}
		err = store.GetBalance(acc)
		if err != nil {
			t.Error(err)
		}
		if math.Abs(initBalance-acc.Balance) > tol {
			t.Error(invalidBalanceValueErr)
		}
	})

	t.Run("DeleteAccount", func(t *testing.T) {
		acc := &models.Account{}
		err := store.InsertAccount(acc)
		if err != nil {
			t.Fatal(err)
		}
		err = store.DeleteAccount(acc)
		if err != nil {
			t.Error(err)
		}
		err = store.GetBalance(acc)
		if err == nil {
			t.Error(invalidBalanceValueErr)
		}
	})

	t.Run("Transfer", func(t *testing.T) {
		accFrom := &models.Account{Balance: 100}
		err := store.InsertAccount(accFrom)
		if err != nil {
			t.Fatal(err)
		}
		accTo := &models.Account{Balance: 10}
		err = store.InsertAccount(accTo)
		if err != nil {
			t.Fatal(err)
		}
		transferInfo, err := store.Transfer(&models.Transfer{
			FromAccountID: accFrom.AccountID,
			ToAccountID:   accTo.AccountID,
			Amount:        90,
		})
		if err != nil {
			t.Error(err)
		}
		if math.Abs(transferInfo.FromAccount.Balance-10) > tol ||
			math.Abs(transferInfo.ToAccount.Balance-100) > tol {
			t.Error(transferCorruptedErr)
		}
		t.Log("New transfer Id: ", transferInfo.TransferID)
	})

	t.Run("TransferNegative", func(t *testing.T) {
		accFrom := &models.Account{Balance: 0.5}
		err := store.InsertAccount(accFrom)
		if err != nil {
			t.Fatal(err)
		}
		accTo := &models.Account{Balance: 10}
		err = store.InsertAccount(accTo)
		if err != nil {
			t.Fatal(err)
		}
		transferInfo, err := store.Transfer(&models.Transfer{
			FromAccountID: accFrom.AccountID,
			ToAccountID:   accTo.AccountID,
			Amount:        11.5,
		})
		if err == nil || transferInfo != nil {
			t.Error(transferCorruptedErr)
		}
	})

	t.Run("Deposit", func(t *testing.T) {
		acc := &models.Account{}
		err := store.InsertAccount(acc)
		if err != nil {
			t.Fatal(err)
		}
		tr := &models.Transfer{
			ToAccountID: acc.AccountID,
			Amount:      42.42,
		}
		accUpdated, err := store.Deposit(tr)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(accUpdated.Balance-tr.Amount) > tol {
			t.Error(invalidBalanceValueErr)
		}
	})

	t.Run("Withdraw", func(t *testing.T) {
		acc := &models.Account{Balance: 100}
		err := store.InsertAccount(acc)
		if err != nil {
			t.Fatal(err)
		}
		tr := &models.Transfer{
			FromAccountID: acc.AccountID,
			Amount:        100,
		}
		accUpdated, err := store.Withdraw(tr)
		if err != nil {
			t.Fatal(err)
		}
		if accUpdated.Balance > tol {
			t.Error(invalidBalanceValueErr)
		}
	})

	t.Run("GetTransfers", func(t *testing.T) {
		var initBalance float64 = 100
		acc := &models.Account{Balance: initBalance}
		err := store.InsertAccount(acc)
		if err != nil {
			t.Fatal(err)
		}
		tr := &models.Transfer{
			FromAccountID: acc.AccountID,
			ToAccountID:   acc.AccountID,
			Amount:        42.5,
		}
		store.Deposit(tr)
		store.Deposit(tr)
		store.Withdraw(tr)

		store.GetBalance(acc)

		transfers, err := store.GetTransfersHistory(&models.TransferHisotoryRequest{
			AccountID: acc.AccountID,
			NDays:     1,
		})
		if err != nil {
			t.Error(err)
		}
		if len(transfers) != 3 {
			t.Error(transferCorruptedErr)
		}
		summ := 0.0
		for _, transfer := range transfers {
			summ += transfer.Amount
		}
		if (summ - acc.Balance) > tol {
			t.Error(transferCorruptedErr)
		}
	})

	t.Run("DropTable", func(t *testing.T) {
		err := store.dropTable("account")
		if err != nil {
			t.Error(err)
		}
	})
}

func TestStoreConcurrentTransfer(t *testing.T) {
	dbPath := "/tmp/tets_concurrent.db"
	store, err := New(dbPath, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	defer os.RemoveAll(dbPath)

	var initBalance float64 = 100
	accFrom := &models.Account{Balance: initBalance}
	store.InsertAccount(accFrom)
	accTo := &models.Account{}
	store.InsertAccount(accTo)

	n := 10
	errs := make(chan error)
	for i := 0; i < n; i++ {
		go func() {
			_, err := store.Transfer(&models.Transfer{
				FromAccountID: accFrom.AccountID,
				ToAccountID:   accTo.AccountID,
				Amount:        2,
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		if err != nil {
			t.Error(err)
		}
	}

	store.GetBalance(accFrom)
	if math.Abs(initBalance-float64(n*2)-accFrom.Balance) > tol {
		t.Fatal(invalidBalanceValueErr)
	}
}
