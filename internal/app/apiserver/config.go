package apiserver

// Config __
type Config struct {
	BindAddr string `toml:"bind_addr"`
	LogLevel string `toml:"log_level"`
}

// NewConfig __
func NewConfig() *Config {
	return &Config{
		BindAddr: ":8010",
		LogLevel: "debug",
	}
}
