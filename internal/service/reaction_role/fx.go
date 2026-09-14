package reactionrole

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	interactioncommand "github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "reactionRole"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewService),
		fx.Provide(
			discord.AsEventListener(NewEventListener),
			interactioncommand.AsCommand(NewReactionRoleCommand),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
