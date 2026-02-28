package pprof

import "go.uber.org/config"

const (
	ConfigKey = "pprof"
)

type Config struct {
	Addr string `yaml:"addr"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{Addr: ":6060"}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
