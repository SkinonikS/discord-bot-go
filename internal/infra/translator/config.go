package translator

import (
	"fmt"

	disgodiscord "github.com/disgoorg/disgo/discord"
	"go.uber.org/config"
)

const (
	ConfigKey = "translator"
)

type Config struct {
	DefaultLocale    disgodiscord.Locale   `yaml:"default_locale"`
	AvailableLocales []disgodiscord.Locale `yaml:"available_locales"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, fmt.Errorf("failed to populate translator config: %w", err)
	}
	return cfg, nil
}
