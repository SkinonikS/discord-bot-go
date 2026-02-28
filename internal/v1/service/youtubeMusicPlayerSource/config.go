package youtubeMusicPlayerSource

import (
	"time"

	"go.uber.org/config"
)

const (
	ConfigKey = "service.youtubeMusicPlayerSource"
)

type Config struct {
	YtdlpPath string        `yaml:"ytdlpPath"`
	Timeout   time.Duration `yaml:"timeout"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
