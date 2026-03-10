package database

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "database"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig, New),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
