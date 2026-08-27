package music_player

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/lavalink"
	"github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "musicPlayer"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig, NewService),
		fx.Provide(
			lavalink.AsEventListener(NewPlayerEventListener),
			interaction_command.AsCommand(NewMusicCommand),
		),
		fx.Invoke(func(disgolink.Client) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
