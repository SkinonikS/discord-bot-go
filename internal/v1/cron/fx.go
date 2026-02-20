package cron

import (
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "cron"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(New),
		fx.Invoke(func(gocron.Scheduler) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
