package tempVoiceChannel

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "tempVoiceChannel"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewHandler, NewInteractionCommand),
		fx.Invoke(func(c *interactionCommand.Registry, cmd *InteractionCommand) error {
			return c.Register(cmd.Name(), cmd)
		}),
		fx.Invoke(func(s *discordgo.Session, handler *Handler) {
			s.AddHandler(handler.Handle)
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
