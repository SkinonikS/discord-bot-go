package interactionCommand

import "go.uber.org/config"

const (
	ConfigKey = "service.interactionCommand"
)

type Config struct {
	OwnerID string `yaml:"ownerID"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
