package logger

import (
	"go.uber.org/config"
	"go.uber.org/zap/zapcore"
)

const (
	ConfigKey = "logger"
)

type Level string

type Config struct {
	Level   Level `yaml:"level"`
	Disable bool  `yaml:"disable"`
}

func NewConfig(provider config.Provider) (*Config, error) {
	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return nil, err
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
