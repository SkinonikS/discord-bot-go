package discord

import (
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/config"
)

const (
	ConfigKey = "discord"
)

type Config struct {
	AppID       snowflake.ID `yaml:"appID"`
	Token       string       `yaml:"token"`
	WorkerCount uint16       `yaml:"workerCount"`
	ShardCount  uint32       `yaml:"shardCount"`
	ShardID     uint32       `yaml:"shardID"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
