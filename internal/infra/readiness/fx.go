package readiness

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/http_server"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "readiness"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewRegistry),
		fx.Provide(
			http_server.AsHandler(NewHTTPHandler),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsHandler(f any) any {
	return fx.Annotate(f, fx.As(new(Handler)), fx.ResultTags(`group:"readiness_handlers"`))
}
