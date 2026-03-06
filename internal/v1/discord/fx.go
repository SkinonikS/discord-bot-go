package discord

import (
	"github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	"github.com/disgoorg/disgo/bot"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "discord"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig, discord.NewUpTime, NewWorkerPool, New),
		fx.Provide(
			AsEventListener(NewEventListener),
		),
		fx.Invoke(
			func(*bot.Client) {},
			func(*discord.WorkerPool) {},
			func(discord.UpTime) {},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsEventListener(f any) any {
	return fx.Annotate(f, fx.As(new(bot.EventListener)), fx.ResultTags(`group:"discord_handlers"`))
}
