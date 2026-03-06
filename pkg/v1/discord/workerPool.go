package discord

import (
	"context"
	"errors"

	disgobot "github.com/disgoorg/disgo/bot"
	"go.uber.org/zap"
)

type WorkerPool struct {
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

func NewWorkerPool(p WorkerPoolParams) *WorkerPool {
	return &WorkerPool{
		eventsChan:     make(chan disgobot.Event, p.WorkerCount*10),
		workers:        p.WorkerCount,
		eventListeners: p.EventListeners,
		log:            p.Log,
	}
}

func (p *WorkerPool) OnEvent(event disgobot.Event) {
	select {
	case p.eventsChan <- event:
	default:
		p.log.Warnw("event bus full, dropped event")
	}
}

func (p *WorkerPool) Start() error {
	if p.cancel != nil {
		return errors.New("worker pool already started")
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	for range p.workers {
		go func() {
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

func (p *WorkerPool) Stop() error {
	if p.cancel == nil {
		return errors.New("worker pool not started")
	}
	p.cancel()
	return nil
}
