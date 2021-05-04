package apiserver

import (
	"encoding/json"
	"net/http"

	"github.com/gasparian/money-transfers-api/internal/app/models"
	"github.com/gasparian/money-transfers-api/internal/app/store"
	"github.com/gasparian/money-transfers-api/internal/app/store/sqlstore"
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
	s.router.HandleFunc("/create-account", s.handleCreateAccount())
	s.router.HandleFunc("/delete-account", s.handleDeleteAccount())
	s.router.HandleFunc("/get-balance", s.handleGetBalance())
	s.router.HandleFunc("/transfer", s.handleTransfer())
	s.router.HandleFunc("/deposit", s.handleDeposit())
	s.router.HandleFunc("/withdraw", s.handleWithdraw())
	s.router.HandleFunc("/get-transfers", s.handleGetTransfers())
}

func (s *APIServer) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func (s *APIServer) handleCreateAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		switch r.Method {
		case "POST":
			var acc models.Account
			err := json.NewDecoder(r.Body).Decode(&acc)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = s.store.InsertAccount(&acc)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(acc)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}

func (s *APIServer) handleDeleteAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		switch r.Method {
		case "POST":
			var acc models.Account
			err := json.NewDecoder(r.Body).Decode(&acc)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = s.store.DeleteAccount(acc.AccountID)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}

func (s *APIServer) handleGetBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		switch r.Method {
		case "POST":
			var acc models.Account
			err := json.NewDecoder(r.Body).Decode(&acc)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			acc.Balance, err = s.store.GetBalance(acc.AccountID)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(acc)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}

func (s *APIServer) handleDeposit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		switch r.Method {
		case "POST":
			var tr models.Transfer
			err := json.NewDecoder(r.Body).Decode(&tr)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = s.store.Deposit(&tr)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}

func (s *APIServer) handleWithdraw() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		switch r.Method {
		case "POST":
			var tr models.Transfer
			err := json.NewDecoder(r.Body).Decode(&tr)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = s.store.Withdraw(&tr)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}

func (s *APIServer) handleTransfer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		switch r.Method {
		case "POST":
			var tr models.Transfer
			err := json.NewDecoder(r.Body).Decode(&tr)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			transferInfo, err := s.store.Transfer(&tr)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(transferInfo)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}

func (s *APIServer) handleGetTransfers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		switch r.Method {
		case "POST":
			var tr models.TransferHisotoryRequest
			err := json.NewDecoder(r.Body).Decode(&tr)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			transfers, err := s.store.GetTransfersHistory(&tr)
			if err != nil {
				s.logger.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(transfers)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}
