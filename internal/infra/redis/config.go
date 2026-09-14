package redis

import (
	"fmt"

	"go.uber.org/config"
)

const (
	ConfigKey = "redis"
)

type Config struct {
	Host        string                       `yaml:"host"`
	Port        int                          `yaml:"port"`
	Username    string                       `yaml:"username,omitempty"`
	Password    string                       `yaml:"password,omitempty"`
	Connections map[string]*ConnectionConfig `yaml:"connections"`
}

type ConnectionConfig struct {
	DB int `yaml:"db"`
}

func newConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, fmt.Errorf("failed to populate redis config: %w", err)
	}
	return cfg, nil
}
