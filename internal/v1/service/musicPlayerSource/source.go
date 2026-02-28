package musicPlayerSource

import (
	"context"
	"io"
	"time"
)

type Track struct {
	Title    string
	URL      string
	Duration time.Duration
}

type Source interface {
	Name() string
	Resolve(ctx context.Context, url string) (*Track, error)
	Stream(ctx context.Context, url string) (io.ReadCloser, error)
}
