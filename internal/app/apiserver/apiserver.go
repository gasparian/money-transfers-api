package apiserver

import (
	"encoding/json"
	"fmt"
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
	s.router.HandleFunc("/deposit", s.handleDeposit())
	s.router.HandleFunc("/withdraw", s.handleWithdraw())
	s.router.HandleFunc("/transfer", s.handleTransfer())
	s.router.HandleFunc("/get-transfers", s.handleGetTransfers())
}

func (s *APIServer) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func (s *APIServer) apiWrapper(f func(w http.ResponseWriter, r *http.Request) error, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	switch r.Method {
	case "POST":
		err := f(w, r)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Method: %s; error: %s", r.URL.Path, err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
	}
}

func (s *APIServer) handleCreateAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.apiWrapper(
			func(w http.ResponseWriter, r *http.Request) error {
				var acc models.Account
				err := json.NewDecoder(r.Body).Decode(&acc)
				if err != nil {
					return err
				}
				err = s.store.InsertAccount(&acc)
				if err != nil {
					return err
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(acc)
				return nil
			},
			w, r,
		)
	}
}

func (s *APIServer) handleDeleteAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.apiWrapper(
			func(w http.ResponseWriter, r *http.Request) error {
				var acc models.Account
				err := json.NewDecoder(r.Body).Decode(&acc)
				if err != nil {
					return err
				}
				err = s.store.DeleteAccount(&acc)
				if err != nil {
					return err
				}
				w.WriteHeader(http.StatusOK)
				return nil
			},
			w, r,
		)
	}
}

func (s *APIServer) handleGetBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.apiWrapper(
			func(w http.ResponseWriter, r *http.Request) error {
				var acc models.Account
				err := json.NewDecoder(r.Body).Decode(&acc)
				if err != nil {
					return err
				}
				err = s.store.GetBalance(&acc)
				if err != nil {
					return err
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(acc)
				return nil
			},
			w, r,
		)
	}
}

func (s *APIServer) handleDeposit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.apiWrapper(
			func(w http.ResponseWriter, r *http.Request) error {
				var tr models.Transfer
				err := json.NewDecoder(r.Body).Decode(&tr)
				if err != nil {
					return err
				}
				acc, err := s.store.Deposit(&tr)
				if err != nil {
					return err
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(acc)
				return nil
			},
			w, r,
		)
	}
}

func (s *APIServer) handleWithdraw() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.apiWrapper(
			func(w http.ResponseWriter, r *http.Request) error {
				var tr models.Transfer
				err := json.NewDecoder(r.Body).Decode(&tr)
				if err != nil {
					return err
				}
				acc, err := s.store.Withdraw(&tr)
				if err != nil {
					return err
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(acc)
				return nil
			},
			w, r,
		)
	}
}

func (s *APIServer) handleTransfer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.apiWrapper(
			func(w http.ResponseWriter, r *http.Request) error {
				var tr models.Transfer
				err := json.NewDecoder(r.Body).Decode(&tr)
				if err != nil {
					return err
				}
				transferInfo, err := s.store.Transfer(&tr)
				if err != nil {
					return err
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(transferInfo)
				return nil
			},
			w, r,
		)
	}
}

func (s *APIServer) handleGetTransfers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.apiWrapper(
			func(w http.ResponseWriter, r *http.Request) error {
				var tr models.TransferHisotoryRequest
				err := json.NewDecoder(r.Body).Decode(&tr)
				if err != nil {
					return err
				}
				transfers, err := s.store.GetTransfersHistory(&tr)
				if err != nil {
					return err
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(transfers)
				return nil
			},
			w, r,
		)
	}
}
