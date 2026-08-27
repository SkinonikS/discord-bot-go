package temp_voice_channel

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	"github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "tempVoiceChannel"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewService, NewChannelRepo, NewChannelStateRepo),
		fx.Provide(
			discord.AsEventListener(NewEventListener),
			interaction_command.AsCommand(NewTempVoiceCommand),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
