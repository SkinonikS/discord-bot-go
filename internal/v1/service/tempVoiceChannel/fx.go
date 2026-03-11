package tempVoiceChannel

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "tempVoiceChannel"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(
			discord.AsEventListener(NewEventListener),
			interactionCommand.AsCommand(NewInteractionCommand),
		),
		fx.Provide(
			NewService,
			NewChannelRepo,
			NewChannelStateRepo,
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
