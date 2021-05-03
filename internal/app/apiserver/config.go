package apiserver

// Config holds needed data to run db and api server
type Config struct {
	BindAddr     string `toml:"bind_addr"`
	LogLevel     string `toml:"log_level"`
	DbPath       string `toml:"db_path"`
	QueryTimeout uint32 `toml:"query_timeout"`
}

// NewConfig instantiates the new configuration object
func NewConfig() *Config {
	return &Config{
		BindAddr:     ":8010",
		LogLevel:     "debug",
		DbPath:       "/tmp/sqlite.db",
		QueryTimeout: 10,
	}
}
