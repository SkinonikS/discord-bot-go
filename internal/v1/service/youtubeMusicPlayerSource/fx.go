package youtubeMusicPlayerSource

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayerSource"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "youtubeMusicPlayerSource"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig),
		fx.Provide(
			musicPlayerSource.AsSource(New),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
