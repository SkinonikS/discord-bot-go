package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "logger"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(newConfig, New),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
