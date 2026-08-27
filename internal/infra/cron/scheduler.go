package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	Jobs []Job `group:"cron_jobs"`
	Log  *zap.Logger
	Lc   fx.Lifecycle
}

func New(p Params) (gocron.Scheduler, error) {
	log := NewLogger(p.Log.Sugar())
	scheduler, err := gocron.NewScheduler(
		gocron.WithLogger(log),
		gocron.WithLocation(time.UTC),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	for _, job := range p.Jobs {
		_, err := scheduler.NewJob(job.Definition(), job.Task(),
			gocron.WithEventListeners(
				gocron.AfterJobRunsWithError(func(id uuid.UUID, name string, err error) {
					p.Log.Error("job failed", zap.String("id", id.String()), zap.String("name", name), zap.Error(err))
				}),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add job to scheduler: %w", err)
		}
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
