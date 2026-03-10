package lavaLink

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "lavaLink"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(New, NewConfig),
		fx.Provide(
			discord.AsEventListener(NewEventListener),
		),
		fx.Invoke(func(disgolink.Client) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsEventListener(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(disgolink.EventListener)),
		fx.ResultTags(`group:"lavaLink_event_listeners"`),
	)
}
