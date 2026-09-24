package logger

import (
	"fmt"

	"go.uber.org/config"
	"go.uber.org/zap/zapcore"
)

const (
	ConfigKey = "logger"
)

type Level string

type Format string

const (
	FormatPretty Format = "pretty"
	FormatJSON   Format = "json"
)

type Config struct {
	Level   Level  `yaml:"level"`
	Disable bool   `yaml:"disable"`
	Format  Format `yaml:"format"`
}

func newConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, fmt.Errorf("failed to populate logger config: %w", err)
	}
	return cfg, nil
}

func (l Level) String() string {
	return string(l)
}

func (l Level) AsZap() zapcore.Level {
	level, err := zapcore.ParseLevel(l.String())
	if err != nil {
		return zapcore.InfoLevel
	}
	return level
}
