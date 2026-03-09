package musicPlayer

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/lavaLink"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "musicPlayer"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig),
		fx.Provide(
			lavaLink.AsEventListener(NewLavaLinkEventListener),
			interactionCommand.AsCommand(NewInteractionCommand),
		),
		fx.Invoke(func(disgolink.Client) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
