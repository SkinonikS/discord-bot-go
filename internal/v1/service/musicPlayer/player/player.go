package player

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type TrackEncoder interface {
	Play(track *Track, event <-chan EventType) (EventType, error)
	OnStop(stopMessage StopMessage)
	OnError(err error)
}

type StatusType int32

const (
	StatusIdle StatusType = iota
	StatusPlaying
	StatusPaused
)

type EventType uint8

const (
	EventStop EventType = iota
	EventSkip
	EventPause
	EventResume
)

type Player struct {
	mu          sync.RWMutex
	status      StatusType
	eventChan   chan EventType
	current     *Track
	queue       *Queue
	loop        *Loop
	log         *zap.SugaredLogger
	idleTimeout time.Duration
}

func New(idleTimeout time.Duration, log *zap.Logger) *Player {
	return &Player{
		loop:        NewLoop(100 * time.Millisecond),
		queue:       NewQueue(),
		status:      StatusIdle,
		eventChan:   make(chan EventType, 8),
		idleTimeout: idleTimeout,
		log:         log.Sugar(),
	}
}

func (p *Player) Current() *Track {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current
}

func (p *Player) Queue() *Queue {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.queue.Snapshot()
}

func (p *Player) PushQueue(track *Track) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queue.Push(track)
}

func (p *Player) ShiftQueue(track *Track) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queue.Shift(track)
}

func (p *Player) Status() StatusType {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *Player) Play(enc TrackEncoder) {
	p.mu.Lock()
	if p.status == StatusPlaying {
		p.mu.Unlock()
		return
	}
	p.status = StatusPlaying
	p.mu.Unlock()

	stopReasonCh := p.loop.Start(func() {
		p.mu.Lock()
		if p.queue.IsEmpty() {
			p.current = nil
			p.mu.Unlock()

			select {
			case <-time.After(p.idleTimeout):
				p.loop.Stop(StopMessage{
					Reason:  StoppedTimeout,
					Message: "Idle timeout reached",
				})
				return
			case evt := <-p.eventChan:
				if evt == EventStop {
					p.loop.Stop(StopMessage{
						Reason:  StoppedManual,
						Message: "Stop event received",
					})
					return
				}
			}
			return
		}

		track := p.queue.Pop()
		p.current = track
		p.mu.Unlock()

		evt, err := enc.Play(track, p.eventChan)
		if err != nil {
			enc.OnError(err)
			return
		}

		if evt == EventStop {
			p.loop.Stop(StopMessage{
				Reason:  StoppedManual,
				Message: "Stop event received",
			})
		}
	})

	go func() {
		reason := <-stopReasonCh
		enc.OnStop(reason)

		p.mu.Lock()
		p.status = StatusIdle
		p.mu.Unlock()
	}()
}

func (p *Player) Stop() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.status == StatusIdle {
		return
	}
	p.sendEvent(EventStop)
}

func (p *Player) Skip() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.status == StatusIdle {
		return
	}
	p.sendEvent(EventSkip)
}

func (p *Player) Pause() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != StatusPlaying {
		return false
	}
	p.status = StatusPaused
	p.sendEvent(EventPause)
	return true
}

func (p *Player) Resume() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != StatusPaused {
		return false
	}
	p.status = StatusPlaying
	p.sendEvent(EventResume)
	return true
}

func (p *Player) Remove(index int) (*Track, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.queue.Remove(index)
}

func (p *Player) sendEvent(e EventType) {
	select {
	case p.eventChan <- e:
	default:
		p.log.Warn("event bus full, dropped event", zap.Uint8("type", uint8(e)))
	}
}
