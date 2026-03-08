package lavaLink

import (
	"time"

	"go.uber.org/config"
)

const (
	ConfigKey = "service.lavaLink"
)

type Config struct {
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	Nodes       []*NodeConfig `yaml:"nodes"`
}

type NodeConfig struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	Secure   bool   `yaml:"secure"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
