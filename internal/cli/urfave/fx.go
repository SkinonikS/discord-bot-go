package urfave

import (
	"github.com/SkinonikS/discord-bot-go/internal/cli/urfave/command"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "urfave"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(New),
		fx.Provide(
			asCommand(command.NewRegisterInteractionCommands),
			asCommand(command.NewLocalInteractionCommands),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func asCommand(f any) any {
	return fx.Annotate(f, fx.ResultTags(`group:"cli_commands"`))
}
