package reactionRole

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "reactionRole"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewHandler, NewInteractionCommand),
		fx.Invoke(func(c *interactionCommand.Registry, cmd *InteractionCommand) error {
			return c.Register(cmd.Name(), cmd)
		}),
		fx.Invoke(func(s *discordgo.Session, handler *Handler) {
			s.AddHandler(handler.HandleAdd)
			s.AddHandler(handler.HandleRemove)
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
