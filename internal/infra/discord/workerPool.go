package discord

import (
	"context"
	"errors"

	disgobot "github.com/disgoorg/disgo/bot"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type workerPoolImpl struct {
	eventsChan     chan disgobot.Event
	workers        uint16
	eventListeners []disgobot.EventListener
	log            *zap.SugaredLogger
	cancel         context.CancelFunc
}

type workerPoolParams struct {
	fx.In

	BotClient      *disgobot.Client
	Lc             fx.Lifecycle
	Log            *zap.Logger
	Config         *Config
	EventListeners []disgobot.EventListener `group:"discord_event_listeners"`
}

func newWorkerPool(p workerPoolParams) *workerPoolImpl {
	wp := &workerPoolImpl{
		eventsChan:     make(chan disgobot.Event, p.Config.WorkerCount*10),
		workers:        p.Config.WorkerCount,
		eventListeners: p.EventListeners,
		log:            p.Log.Sugar(),
	}

	p.BotClient.AddEventListeners(wp)
	p.Lc.Append(fx.StartStopHook(
		func(context.Context) error {
			//nolint:contextcheck,gosec
			if err := wp.Start(); err != nil {
				return err
			}
			p.Log.Info("worker pool started")
			return nil
		},
		func(context.Context) error {
			if err := wp.Stop(); err != nil {
				return err
			}
			p.Log.Info("worker pool stopped")
			return nil
		},
	))

	return wp
}

func (p *workerPoolImpl) OnEvent(event disgobot.Event) {
	select {
	case p.eventsChan <- event:
	default:
		p.log.Warnw("event bus full, dropped event")
	}
}

func (p *workerPoolImpl) Start() error {
	if p.cancel != nil {
		return errors.New("worker pool already started")
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	for range p.workers {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					p.log.Errorf("worker panic recovered: %v", r)
				}
			}()

			for {
				select {
				case event := <-p.eventsChan:
					for _, el := range p.eventListeners {
						el.OnEvent(event)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return nil
}

func (p *workerPoolImpl) Stop() error {
	if p.cancel == nil {
		return errors.New("worker pool not started")
	}
	p.cancel()
	return nil
}
