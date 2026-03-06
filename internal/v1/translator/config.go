package translator

import (
	disgodiscord "github.com/disgoorg/disgo/discord"
	"go.uber.org/config"
)

const (
	ConfigKey = "translator"
)

type Config struct {
	DefaultLocale    disgodiscord.Locale   `yaml:"defaultLocale"`
	AvailableLocales []disgodiscord.Locale `yaml:"availableLocales"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
