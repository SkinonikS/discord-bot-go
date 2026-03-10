package discord

import (
	"context"
	"errors"

	disgobot "github.com/disgoorg/disgo/bot"
	"go.uber.org/zap"
)

type WorkerPool interface {
	disgobot.EventListener
	Start() error
	Stop() error
}

type workerPoolImpl struct {
	eventsChan     chan disgobot.Event
	workers        uint16
	eventListeners []disgobot.EventListener
	log            *zap.SugaredLogger
	cancel         context.CancelFunc
}

type WorkerPoolParams struct {
	WorkerCount    uint16
	EventListeners []disgobot.EventListener
	Log            *zap.SugaredLogger
}

func NewWorkerPool(p WorkerPoolParams) WorkerPool {
	return &workerPoolImpl{
		eventsChan:     make(chan disgobot.Event, p.WorkerCount*10),
		workers:        p.WorkerCount,
		eventListeners: p.EventListeners,
		log:            p.Log,
	}
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
