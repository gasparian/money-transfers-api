package apiserver

import (
	"net/http"

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

// Start run db and api server
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
}

func (s *APIServer) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
