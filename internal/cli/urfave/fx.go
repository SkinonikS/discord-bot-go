package urfave

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "urfave"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(New),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func asCommand(f any) any {
	return fx.Annotate(f, fx.ResultTags(`group:"cli_commands"`))
}
