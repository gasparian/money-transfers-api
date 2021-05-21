package apiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gasparian/money-transfers-api/internal/app/models"
	"github.com/gasparian/money-transfers-api/internal/app/store/sqlstore"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

var (
	badStatusCodeErr     = errors.New("Bad status code")
	wrongAnswerErr       = errors.New("Wrong answer")
	accountNotDeletedErr = errors.New("Account has not been deleted")
)

func addQueryParams(req *http.Request, params map[string]string) {
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
}

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
		initBalance := MoneyAmountJsonView{Integer: 100}
		b, err := json.Marshal(initBalance)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBuffer(b))
		s.handleAccounts().ServeHTTP(rec, req)
		if rec.Code > 204 {
			t.Error(badStatusCodeErr)
		}
		accId := AccountIDJsonView{}
		if err := json.NewDecoder(rec.Body).Decode(&accId); err != nil {
			t.Error(err)
		}
		if accId.ID == 0 {
			t.Error(wrongAnswerErr)
		}
	})

	t.Run("DeleteAccount", func(t *testing.T) {
		rec := httptest.NewRecorder()
		acc, err := store.InsertAccount(models.MoneyAmount{Integer: 100})
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/accounts", nil)
		addQueryParams(req, map[string]string{
			"account_id": fmt.Sprintf("%v", acc.AccountID),
		})
		s.handleAccounts().ServeHTTP(rec, req)
		if rec.Code > 204 {
			t.Error(badStatusCodeErr)
		}
		_, err = store.GetAccount(acc.AccountID)
		if err == nil {
			t.Error(accountNotDeletedErr)
		}
	})

	t.Run("GetAccount", func(t *testing.T) {
		rec := httptest.NewRecorder()
		initBalance := models.MoneyAmount{Integer: 100}
		acc, err := store.InsertAccount(initBalance)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
		addQueryParams(req, map[string]string{
			"account_id": fmt.Sprintf("%v", acc.AccountID),
		})
		s.handleAccounts().ServeHTTP(rec, req)
		if rec.Code > 204 {
			t.Error(badStatusCodeErr)
		}
		if err := json.NewDecoder(rec.Body).Decode(&acc); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(&acc.Balance, &initBalance) {
			t.Error(wrongAnswerErr)
		}
	})

	t.Run("Transfer", func(t *testing.T) {
		rec := httptest.NewRecorder()
		accFromInitBalance := models.MoneyAmount{Integer: 100}
		accFrom, err := store.InsertAccount(accFromInitBalance)
		if err != nil {
			t.Fatal(err)
		}
		accToInitBalance := models.MoneyAmount{}
		accTo, err := store.InsertAccount(accToInitBalance)
		if err != nil {
			t.Fatal(err)
		}
		tr := TransactionJsonView{
			FromAccountID: accFrom.AccountID,
			ToAccountID:   accTo.AccountID,
			Amount:        MoneyAmountJsonView(accFromInitBalance),
		}
		b, err := json.Marshal(tr)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/transfer-money", bytes.NewBuffer(b))
		s.handleTransferMoney().ServeHTTP(rec, req)
		if rec.Code > 204 {
			t.Error(badStatusCodeErr)
		}
		accFromNew, _ := store.GetAccount(accFrom.AccountID)
		accToNew, _ := store.GetAccount(accTo.AccountID)
		if models.CompareMoney(&accFromNew.Balance, &accFromInitBalance) >= 0 &&
			models.CompareMoney(&accToNew.Balance, &accToInitBalance) <= 0 &&
			models.CompareMoney(&accFromNew.Balance, &accToNew.Balance) >= 0 {
			t.Error(wrongAnswerErr)
		}
	})

	t.Run("GetTransactions", func(t *testing.T) {
		rec := httptest.NewRecorder()
		accFromInitBalance := models.MoneyAmount{Integer: 100}
		accFrom, err := store.InsertAccount(accFromInitBalance)
		if err != nil {
			t.Fatal(err)
		}
		accToInitBalance := models.MoneyAmount{}
		accTo, err := store.InsertAccount(accToInitBalance)
		if err != nil {
			t.Fatal(err)
		}
		transferMoneyAmount := models.MoneyAmount{Integer: 20}
		nTransfers := 5
		for i := 0; i < nTransfers; i++ {
			store.TransferMoney(
				accTo.AccountID,
				accFrom.AccountID,
				transferMoneyAmount,
			)
		}

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
		addQueryParams(req, map[string]string{
			"account_id":  fmt.Sprintf("%v", accFrom.AccountID),
			"n_last_days": fmt.Sprintf("%v", 1),
			"limit":       fmt.Sprintf("%v", 3),
		})
		s.handleTransactions().ServeHTTP(rec, req)
		if rec.Code > 204 {
			t.Error(badStatusCodeErr)
		}
		transactions := make([]TransactionJsonView, 3)
		if err := json.NewDecoder(rec.Body).Decode(&transactions); err != nil {
			t.Error(err)
		}
		for _, tr := range transactions {
			if tr.Amount.Integer != 20 || tr.FromAccountID != accFrom.AccountID {
				t.Fatal(wrongAnswerErr)
			}
		}
		if len(transactions) != 3 {
			t.Error(wrongAnswerErr)
		}
	})
}
