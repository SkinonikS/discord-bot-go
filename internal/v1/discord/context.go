package discord

import (
	"context"
	"time"
)

const (
	HandlerContextTimeout = 2 * time.Second
)

func DefaultHandlerContext() (context.Context, context.CancelFunc) {
	return HandlerContext(context.Background())
}

func HandlerContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, HandlerContextTimeout)
}
