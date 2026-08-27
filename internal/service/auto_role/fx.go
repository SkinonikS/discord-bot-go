package auto_role

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	"github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "autoRole"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewService, NewRepo),
		fx.Provide(
			discord.AsEventListener(NewEventListener),
			interaction_command.AsCommand(NewAutoRoleCommand),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
