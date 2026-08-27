package internal

import (
	"github.com/SkinonikS/discord-bot-go/internal/app/cli/internal/cli"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "cli"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		cli.NewModule(),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
