package musicplayer

import (
	"time"

	"go.uber.org/config"
)

const (
	ConfigKey = "music_player"
)

type Config struct {
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
