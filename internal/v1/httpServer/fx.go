package httpServer

import (
	"github.com/gin-contrib/graceful"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "httpServer"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(New, NewConfig),
		fx.Invoke(func(*graceful.Graceful) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsHandler(f any) any {
	return fx.Annotate(f, fx.As(new(Handler)), fx.ResultTags(`group:"http_handlers"`))
}
