package redis

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "redis"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(newConfig, NewManager),
		fx.Invoke(func(Manager) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
