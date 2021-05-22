package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gasparian/money-transfers-api/internal/app/store"
	"github.com/gasparian/money-transfers-api/internal/app/store/sqlstore"
)

var (
	idNotPresented        = errors.New("Account id not presented in request params")
	timeRangeNotPresented = errors.New("Number of days to query transfers stats is not presented in request params")
	limitNotPresented     = errors.New("Query limit is not presented in reqeust params")
)

// APIServer holds data needed to run api server
type APIServer struct {
	config *Config
	logger *logger
	router *http.ServeMux
	store  store.Store
}

// New creates new instance of APIServer struct
func New(config *Config) *APIServer {
	s := &APIServer{
		config: config,
		logger: NewLogger(),
		router: http.NewServeMux(),
	}
	s.configureLogger()
	s.configureRouter()
	return s
}

func (s *APIServer) setStore(store store.Store) {
	s.store = store
}

// Start runs db and api server
func (s *APIServer) Start() error {
	store, err := sqlstore.New(s.config.DbPath, s.config.QueryTimeout)
	if err != nil {
		return nil
	}
	defer store.Close()
	s.setStore(store)
	s.logger.Info("Starting api server")
	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *APIServer) configureLogger() {
	s.logger.SetLevel(s.config.LogLevel)
}

func (s *APIServer) configureRouter() {
	s.router.HandleFunc("/health", s.handleHealth())
	s.router.HandleFunc("/api/v1/accounts", s.handleAccounts())
	s.router.HandleFunc("/api/v1/transfer-money", s.handleTransferMoney())
	s.router.HandleFunc("/api/v1/transactions", s.handleTransactions())
}

func (s *APIServer) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func (s *APIServer) handleError(err error, statusCode int, w http.ResponseWriter, r *http.Request) {
	s.logger.Error(fmt.Sprintf("Method: %s; error: %s", r.URL.Path, err.Error()))
	w.WriteHeader(statusCode)
}

func parseIntQueryParams(r *http.Request, paramNames ...string) (map[string]int64, error) {
	params := r.URL.Query()
	m := make(map[string]int64)
	for _, param := range paramNames {
		val := params.Get(param)
		conv, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		m[param] = conv
	}
	return m, nil
}

func (s *APIServer) handleAccounts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.Header().Set("Content-type", "application/json")
			var acc AccountJsonView
			err := json.NewDecoder(r.Body).Decode(&acc)
			if err != nil {
				s.handleError(err, http.StatusBadRequest, w, r)
				return
			}
			accModel, err := s.store.InsertAccount(acc.Balance)
			if err != nil {
				s.handleError(err, http.StatusInternalServerError, w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AccountIDJsonView{ID: accModel.AccountID})
		case "DELETE":
			valMap, err := parseIntQueryParams(r, "account_id")
			if err != nil {
				s.handleError(err, http.StatusBadRequest, w, r)
				return
			}
			err = s.store.DeleteAccount(valMap["account_id"])
			if err != nil {
				s.handleError(err, http.StatusInternalServerError, w, r)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case "GET":
			w.Header().Set("Content-type", "application/json")
			valMap, err := parseIntQueryParams(r, "account_id")
			if err != nil {
				s.handleError(err, http.StatusBadRequest, w, r)
				return
			}
			accModel, err := s.store.GetAccount(valMap["account_id"])
			if err != nil {
				s.handleError(err, http.StatusInternalServerError, w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AccountJsonView{
				AccountID: accModel.AccountID,
				Balance:   accModel.Balance,
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}

func (s *APIServer) handleTransferMoney() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var tr TransactionJsonView
			err := json.NewDecoder(r.Body).Decode(&tr)
			if err != nil {
				s.handleError(err, http.StatusBadRequest, w, r)
			}
			err = s.store.TransferMoney(
				tr.ToAccountID,
				tr.FromAccountID,
				tr.Amount,
			)
			if err != nil {
				s.handleError(err, http.StatusInternalServerError, w, r)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}

func (s *APIServer) handleTransactions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-type", "application/json")
			valMap, err := parseIntQueryParams(r, "account_id", "n_last_days", "limit")
			if err != nil {
				s.handleError(err, http.StatusBadRequest, w, r)
				return
			}

			transactions, err := s.store.GetTransactionsHistory(
				valMap["account_id"],
				valMap["n_last_days"],
				valMap["limit"],
			)
			if err != nil {
				s.handleError(err, http.StatusInternalServerError, w, r)
				return
			}
			transactionsJson := make([]TransactionJsonView, len(transactions))
			for i, tr := range transactions {
				transactionsJson[i] = TransactionJsonView{
					Timestamp:     tr.Timestamp,
					FromAccountID: tr.FromAccountID,
					ToAccountID:   tr.ToAccountID,
					Amount:        tr.Amount,
				}
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(transactionsJson)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}
