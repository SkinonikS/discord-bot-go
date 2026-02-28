package musicPlayerSource

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "musicPlayerSource"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewRegistry),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsSource(f any) any {
	return fx.Annotate(f, fx.As(new(Source)), fx.ResultTags(`group:"music_player_sources"`))
}
