package cache

import (
	"fmt"

	"go.uber.org/config"
)

const (
	ConfigKey = "cache"
)

type Config struct {
	Stores map[string]*StoreConfig `yaml:"clients"`
}

type StoreConfig struct {
	Driver string         `yaml:"driver"`
	Config map[string]any `yaml:"config,omitempty"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{
		Stores: map[string]*StoreConfig{},
	}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, fmt.Errorf("failed to populate cache config: %w", err)
	}
	return cfg, nil
}
