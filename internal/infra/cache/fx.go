package cache

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/cache/driver"
	"github.com/SkinonikS/discord-bot-go/internal/infra/cache/driver/memory"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const ModuleName = "cache"

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewManager, NewConfig),
		fx.Provide(
			AsDriverFactory(memory.NewFactory),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsDriverFactory(f any) any {
	return fx.Annotate(f, fx.As(new(driver.Factory)), fx.ResultTags(`group:"cache_driver_factories"`))
}
