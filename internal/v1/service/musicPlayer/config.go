package musicPlayer

import (
	"time"

	"go.uber.org/config"
)

const (
	ConfigKey = "service.musicPlayer"
)

type Config struct {
	FfmpegPath  string        `yaml:"ffmpegPath"`
	IdleTimeout time.Duration `yaml:"idleTimeout"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
