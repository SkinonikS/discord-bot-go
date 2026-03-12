package reactionRole

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "reactionRole"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewService, NewRepo),
		fx.Provide(
			discord.AsEventListener(NewEventListener),
			interactionCommand.AsCommand(NewReactionRoleCommand),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
