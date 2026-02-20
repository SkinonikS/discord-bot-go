package discord

import (
	"context"
	"time"
)

const (
	HandlerContextTimeout = 2 * time.Second
)

func DefaultHandlerContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), HandlerContextTimeout)
}

func HandlerContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, HandlerContextTimeout)
}
