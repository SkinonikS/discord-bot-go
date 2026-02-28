package musicPlayer

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
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
			interactionCommand.AsCommand(NewInteractionCommand),
		),
		fx.Invoke(func(session *discordgo.Session, handler *Handler) {
			session.AddHandler(handler.Handle)
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
