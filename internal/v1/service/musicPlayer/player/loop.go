package player

import (
	"sync"
	"time"
)

type StopReason uint8

const (
	StoppedManual StopReason = iota
	StoppedTimeout
)

type StopMessage struct {
	Reason  StopReason
	Message any
}

type Loop struct {
	stopCh       chan struct{}
	doneCh       chan StopMessage
	running      bool
	reason       StopMessage
	mu           sync.Mutex
	waitInterval time.Duration
}

func NewLoop(waitInterval time.Duration) *Loop {
	return &Loop{
		waitInterval: waitInterval,
	}
}

func (l *Loop) Start(cb func()) <-chan StopMessage {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.running {
		return l.doneCh
	}

	l.stopCh = make(chan struct{})
	l.doneCh = make(chan StopMessage, 1)
	l.running = true

	go func() {
		defer func() {
			l.mu.Lock()
			l.running = false
			l.mu.Unlock()
			l.doneCh <- l.reason
			close(l.doneCh)
		}()

		for {
			select {
			case <-l.stopCh:
				return
			default:
				cb()
				time.Sleep(l.waitInterval)
			}
		}
	}()

	return l.doneCh
}

func (l *Loop) Stop(reason StopMessage) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.running {
		l.reason = reason
		close(l.stopCh)
	}
}
