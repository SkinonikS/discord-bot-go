package config

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "config"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(New),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
