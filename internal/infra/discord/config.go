package discord

import (
	"fmt"

	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/config"
)

const (
	ConfigKey = "discord"
)

type Config struct {
	AppID       snowflake.ID `yaml:"app_id"`
	Token       string       `yaml:"token"`
	WorkerCount uint16       `yaml:"worker_count"`
	ShardCount  int          `yaml:"shard_count"`
	ShardID     int          `yaml:"shard_id"`
}

func newConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(&cfg); err != nil {
		return nil, fmt.Errorf("failed to populate discord config: %w", err)
	}
	return cfg, nil
}
