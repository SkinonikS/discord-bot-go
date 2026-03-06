package discord

import (
	"context"
	"time"
)

const (
	HandlerContextTimeout = 2 * time.Second
)

func DefaultEventListenContext() (context.Context, context.CancelFunc) {
	return EventListenerContext(context.Background())
}

func EventListenerContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, HandlerContextTimeout)
}
