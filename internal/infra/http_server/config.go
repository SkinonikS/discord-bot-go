package http_server

import (
	"fmt"

	"go.uber.org/config"
)

const (
	ConfigKey = "http_server"
)

type Config struct {
	Addr string `yaml:"addr"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, fmt.Errorf("failed to populate http server config: %w", err)
	}
	return cfg, nil
}
