package cron

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In
	Log *zap.Logger
	Lc  fx.Lifecycle
}

func New(p Params) (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	p.Lc.Append(fx.StartStopHook(
		func(context.Context) error {
			go scheduler.Start()
			p.Log.Info("cron scheduler started")
			return nil
		},
		func(context.Context) error {
			if err := scheduler.Shutdown(); err != nil {
				return fmt.Errorf("failed to shutdown scheduler: %w", err)
			}

			p.Log.Info("cron scheduler stopped")
			return nil
		},
	))

	return scheduler, nil
}
