package lavaLink

import (
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

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
