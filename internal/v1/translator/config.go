package translator

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/config"
)

const (
	ConfigKey = "translator"
)

type Config struct {
	DefaultLocale    discordgo.Locale   `yaml:"defaultLocale"`
	AvailableLocales []discordgo.Locale `yaml:"availableLocales"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
