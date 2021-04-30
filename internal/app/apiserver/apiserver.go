package apiserver

// APIServer __
type APIServer struct {
	config *Config
	logger *logger
}

// New __
func New(config *Config) *APIServer {
	return &APIServer{
		config: config,
		logger: NewLogger(),
	}
}

// Start __
func (s *APIServer) Start() error {
	s.configureLogger()
	s.logger.Info("Starting api server")
	return nil
}

func (s *APIServer) configureLogger() {
	s.logger.SetLevel(s.config.LogLevel)
}
