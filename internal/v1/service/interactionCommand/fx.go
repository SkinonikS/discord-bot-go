package interactionCommand

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand/command"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "interactionCommand"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewHandler, NewRegistry),
		fx.Provide(
			AsCommand(command.NewPing),
			AsCommand(command.NewInfo),
		),
		fx.Invoke(func(s *discordgo.Session, handler *Handler) {
			s.AddHandler(handler.Handle)
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsCommand(f any) any {
	return fx.Annotate(f, fx.As(new(command.Command)), fx.ResultTags(`group:"discord_commands"`))
}
