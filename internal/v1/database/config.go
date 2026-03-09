package database

import "go.uber.org/config"

const (
	ConfigKey = "database"
)

type SqliteConfig struct {
	Path string `yaml:"path"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
	SSLMode  string `yaml:"sslMode"`
}

type Config struct {
	Driver   string          `yaml:"driver"`
	Postgres *PostgresConfig `yaml:"postgres" optional:"true"`
	Sqlite   *SqliteConfig   `yaml:"sqlite" optional:"true"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	var cfg *Config
	if err := provider.Get(ConfigKey).Populate(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
