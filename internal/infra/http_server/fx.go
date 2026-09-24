package httpserver

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "httpServer"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(New, newConfig),
		fx.Invoke(func(*fiber.App) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsHandler(f any) any {
	return fx.Annotate(f, fx.As(new(Handler)), fx.ResultTags(`group:"http_handlers"`))
}
