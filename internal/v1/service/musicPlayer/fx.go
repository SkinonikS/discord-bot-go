package musicPlayer

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "musicPlayer"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig, NewManager, NewHandler),
		fx.Provide(
		// TODO: Until DAVE encryption is added.
		//interactionCommand.AsCommand(NewInteractionCommand),
		),
		fx.Invoke(func(session *discordgo.Session, handler *Handler) {
			session.AddHandler(handler.Handle)
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
