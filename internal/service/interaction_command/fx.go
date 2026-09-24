package interactioncommand

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/cli"
	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "interactionCommand"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewRegistry, NewService),
		fx.Provide(
			AsCommand(NewDiscordManageCommand),
			cli.AsCommand(NewCLICommandsCommand),
			discord.AsEventListener(NewDiscordEventListener),
		),
		fx.Invoke(populateRegistry),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsCommand(f any) any {
	return fx.Annotate(f, fx.As(new(Command)), fx.ResultTags(`group:"discord_commands"`))
}
