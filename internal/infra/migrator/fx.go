package migrator

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const ModuleName = "migrator"

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewRegistry),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
