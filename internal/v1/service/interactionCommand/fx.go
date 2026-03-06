package interactionCommand

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand/command"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "interactionCommand"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewRegistry),
		fx.Provide(
			AsCommand(command.NewPing),
			AsCommand(command.NewInfo),
		),
		fx.Provide(
			discord.AsEventListener(NewEventListener),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsCommand(f any) any {
	return fx.Annotate(f, fx.As(new(Command)), fx.ResultTags(`group:"discord_commands"`))
}
