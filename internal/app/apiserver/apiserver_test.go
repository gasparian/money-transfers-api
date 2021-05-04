package apiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gasparian/money-transfers-api/internal/app/models"
	"github.com/gasparian/money-transfers-api/internal/app/store/sqlstore"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	tol = 1e-4
)

var (
	badStatusCodeErr = errors.New("Bad status code")
	wrongAnswerErr   = errors.New("Wrong answer")
)

func TestAPIServer(t *testing.T) {
	dbPath := "/tmp/tets.db"
	store, err := sqlstore.New(dbPath, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	defer os.RemoveAll(dbPath)

	s := New(NewConfig())
	s.setStore(store)

	t.Run("Health", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		s.handleHealth().ServeHTTP(rec, req)
		if rec.Body.String() != "OK" {
			t.Error("Healthcheck failed")
		}
	})

	t.Run("CreateAccount", func(t *testing.T) {
		rec := httptest.NewRecorder()
		acc := models.Account{Balance: 100}
		b, err := json.Marshal(acc)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/create-account", bytes.NewBuffer(b))
		s.handleCreateAccount().ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Error(badStatusCodeErr)
		}
		if err := json.NewDecoder(rec.Body).Decode(&acc); err != nil {
			t.Error(err)
		}
		if acc.AccountID == 0 {
			t.Error(wrongAnswerErr)
		}
	})

	t.Run("DeleteAccount", func(t *testing.T) {
		rec := httptest.NewRecorder()
		acc := models.Account{}
		store.InsertAccount(&acc)
		b, err := json.Marshal(acc)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/delete-account", bytes.NewBuffer(b))
		s.handleDeleteAccount().ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Error(badStatusCodeErr)
		}
		_, err = store.GetBalance(acc.AccountID)
		if err == nil {
			t.Error()
		}
	})

	t.Run("GetBalance", func(t *testing.T) {
		rec := httptest.NewRecorder()
		acc := models.Account{Balance: 100}
		store.InsertAccount(&acc)
		b, err := json.Marshal(acc)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/get-balance", bytes.NewBuffer(b))
		s.handleGetBalance().ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Error(badStatusCodeErr)
		}
		if err := json.NewDecoder(rec.Body).Decode(&acc); err != nil {
			t.Error(err)
		}
		if math.Abs(acc.Balance-100) > tol {
			t.Error(wrongAnswerErr)
		}
	})

	t.Run("Deposit", func(t *testing.T) {
		rec := httptest.NewRecorder()
		accTo := models.Account{}
		store.InsertAccount(&accTo)
		tr := models.Transfer{
			ToAccountID: accTo.AccountID,
			Amount:      100,
		}
		b, err := json.Marshal(tr)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/deposit", bytes.NewBuffer(b))
		s.handleDeposit().ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Error(badStatusCodeErr)
		}
		if err := json.NewDecoder(rec.Body).Decode(&accTo); err != nil {
			t.Error(err)
		}
		if math.Abs(accTo.Balance-100) > tol {
			t.Error(wrongAnswerErr)
		}
	})

	t.Run("Withdraw", func(t *testing.T) {
		rec := httptest.NewRecorder()
		accFrom := models.Account{Balance: 100}
		store.InsertAccount(&accFrom)
		tr := models.Transfer{
			FromAccountID: accFrom.AccountID,
			Amount:        100,
		}
		b, err := json.Marshal(tr)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/withdraw", bytes.NewBuffer(b))
		s.handleWithdraw().ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Error(badStatusCodeErr)
		}
		if err := json.NewDecoder(rec.Body).Decode(&accFrom); err != nil {
			t.Error(err)
		}
		if accFrom.Balance > tol {
			t.Error(wrongAnswerErr)
		}
	})

	t.Run("Transfer", func(t *testing.T) {
		rec := httptest.NewRecorder()
		accFrom := models.Account{Balance: 100}
		store.InsertAccount(&accFrom)
		accTo := models.Account{}
		store.InsertAccount(&accTo)
		tr := models.Transfer{
			FromAccountID: accFrom.AccountID,
			ToAccountID:   accTo.AccountID,
			Amount:        100,
		}
		b, err := json.Marshal(tr)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/transfer", bytes.NewBuffer(b))
		s.handleTransfer().ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Error(badStatusCodeErr)
		}
		var transferInfo models.TransferResult
		if err := json.NewDecoder(rec.Body).Decode(&transferInfo); err != nil {
			t.Error(err)
		}
		if transferInfo.ToAccountIDBalance <= transferInfo.FromAccountIDBalance {
			t.Error(wrongAnswerErr)
		}
	})

	t.Run("GetTransfers", func(t *testing.T) {
		rec := httptest.NewRecorder()
		acc := models.Account{}
		store.InsertAccount(&acc)
		store.Deposit(&models.Transfer{
			ToAccountID: acc.AccountID,
			Amount:      10,
		})
		store.Deposit(&models.Transfer{
			ToAccountID: acc.AccountID,
			Amount:      10,
		})
		store.Withdraw(&models.Transfer{
			FromAccountID: acc.AccountID,
			Amount:        20,
		})
		tr := models.TransferHisotoryRequest{
			AccountID: acc.AccountID,
			NDays:     1,
		}
		b, err := json.Marshal(tr)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/get-transfers", bytes.NewBuffer(b))
		s.handleGetTransfers().ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Error(badStatusCodeErr)
		}
		transfers := make([]models.Transfer, 0)
		if err := json.NewDecoder(rec.Body).Decode(&transfers); err != nil {
			t.Error(err)
		}
		if len(transfers) != 3 {
			t.Error(wrongAnswerErr)
		}
	})
}
