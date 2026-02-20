package discord

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "discord"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig, New),
		fx.Invoke(func(*discordgo.Session) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
