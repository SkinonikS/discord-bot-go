package ping

import (
	interactioncommand "github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "ping"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(
			interactioncommand.AsCommand(NewDiscordPingCommand),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
