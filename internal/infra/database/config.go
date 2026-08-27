package database

import (
	"fmt"

	"go.uber.org/config"
)

const (
	ConfigKey = "database"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
	SSLMode  string `yaml:"ssl_mode"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	var cfg *Config
	if err := provider.Get(ConfigKey).Populate(&cfg); err != nil {
		return nil, fmt.Errorf("failed to populate database config: %w", err)
	}
	return cfg, nil
}
