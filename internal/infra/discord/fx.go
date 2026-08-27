package discord

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/readiness"
	disgobot "github.com/disgoorg/disgo/bot"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "discord"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig, NewUpTime, New, newWorkerPool),
		fx.Provide(
			AsEventListener(NewEventListener),
			readiness.AsHandler(NewReadinessHandler),
		),
		fx.Invoke(
			func(*disgobot.Client) {},
			func(*workerPoolImpl) {},
			func(UpTime) {},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsEventListener(f any) any {
	return fx.Annotate(f, fx.As(new(disgobot.EventListener)), fx.ResultTags(`group:"discord_event_listeners"`))
}
