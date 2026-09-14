package lavalink

import (
	"fmt"

	"go.uber.org/config"
)

const (
	ConfigKey = "lavaLink"
)

type Config struct {
	Nodes []*NodeConfig `yaml:"nodes"`
}

type NodeConfig struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Secure   bool   `yaml:"secure"`
}

func newConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, fmt.Errorf("failed to populate lavalink config: %w", err)
	}
	return cfg, nil
}
