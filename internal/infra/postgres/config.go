package postgres

import (
	"fmt"

	"go.uber.org/config"
)

const (
	ConfigKey = "postgres"
)

type Config struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`
	DB       string `yaml:"db"`
	SSLMode  string `yaml:"ssl_mode"`
}

func newConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, fmt.Errorf("failed to populate postgres config: %w", err)
	}
	return cfg, nil
}
