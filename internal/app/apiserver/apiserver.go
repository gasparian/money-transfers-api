package apiserver

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

// APIServer __
type APIServer struct {
	config *Config
	logger *logger
	router *http.ServeMux
}

// New __
func New(config *Config) *APIServer {
	return &APIServer{
		config: config,
		logger: NewLogger(),
		router: http.NewServeMux(),
	}
}

// Start __
func (s *APIServer) Start() error {
	s.configureLogger()
	s.configureRouter()
	s.logger.Info("Starting api server")
	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func newDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
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
