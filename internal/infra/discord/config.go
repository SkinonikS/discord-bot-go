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
	ShardCount  uint32       `yaml:"shard_count"`
	ShardID     uint32       `yaml:"shard_id"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(&cfg); err != nil {
		return nil, fmt.Errorf("failed to populate discord config: %w", err)
	}
	return cfg, nil
}
