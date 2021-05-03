package sqlstore

import (
	"errors"
	"github.com/gasparian/money-transfers-api/internal/app/models"
	"os"
	"testing"
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
		acc := &models.Account{Balance: 10}
		err := store.InsertAccount(acc)
		if err != nil {
			t.Fatal(err)
		}
		balance, err := store.GetBalance(acc.AccountID)
		if err != nil {
			t.Error(err)
		}
		if balance != acc.Balance {
			t.Error(invalidBalanceValueErr)
		}
	})

	t.Run("DeleteAccount", func(t *testing.T) {
		acc := &models.Account{}
		err := store.InsertAccount(acc)
		if err != nil {
			t.Fatal(err)
		}
		err = store.DeleteAccount(acc.AccountID)
		if err != nil {
			t.Error(err)
		}
		_, err = store.GetBalance(acc.AccountID)
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
		if !(transferInfo.FromAccountIDBalance < 11 && transferInfo.ToAccountIDBalance > 99) {
			t.Error(transferCorruptedErr)
		}
		t.Log("New transfer Id: ", transferInfo.Transfer.TransferID)
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
		err = store.Deposit(tr)
		if err != nil {
			t.Fatal(err)
		}
		balance, err := store.GetBalance(acc.AccountID)
		if err != nil {
			t.Error(err)
		}
		if balance < tr.Amount {
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
		err = store.Withdraw(tr)
		if err != nil {
			t.Fatal(err)
		}
		balance, err := store.GetBalance(acc.AccountID)
		if err != nil {
			t.Error(err)
		}
		if balance > 1.0 {
			t.Error(invalidBalanceValueErr)
		}
	})

	t.Run("GetTransfers", func(t *testing.T) {
		acc := &models.Account{Balance: 100}
		err := store.InsertAccount(acc)
		if err != nil {
			t.Fatal(err)
		}
		tr := &models.Transfer{
			FromAccountID: acc.AccountID,
			ToAccountID:   acc.AccountID,
			Amount:        42.42,
		}
		store.Deposit(tr)
		store.Deposit(tr)
		store.Withdraw(tr)

		transfers, err := store.GetTransfersHistory(acc.AccountID)
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
		if summ < 42 {
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

	accFrom := &models.Account{Balance: 100}
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

	balance, _ := store.GetBalance(accFrom.AccountID)
	if (accFrom.Balance - float64(n*2) - balance) > 1e-3 {
		t.Fatal(invalidBalanceValueErr)
	}

	for i := 0; i < n; i++ {
		err := <-errs
		if err != nil {
			t.Error(err)
		}
	}
}
