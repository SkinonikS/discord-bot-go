package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	Config *Config
	Lc     fx.Lifecycle
	Log    *zap.Logger
}

func New(p Params) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(p.Config.User),
		url.QueryEscape(p.Config.Password),
		p.Config.Host,
		p.Config.Port,
		p.Config.DB,
		p.Config.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres dsn: %w", err)
	}
	poolConfig.ShouldPing = func(_ context.Context, _ pgxpool.ShouldPingParams) bool {
		return false
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	p.Lc.Append(fx.StartStopHook(func(ctx context.Context) error {
		go func() {
			if err := pool.Ping(ctx); err != nil {
				p.Log.Error("failed to ping postgres database", zap.Error(err))
			} else {
				addr := net.JoinHostPort(p.Config.Host, strconv.Itoa(int(p.Config.Port)))
				p.Log.Info("connection established", zap.String("addr", addr))
			}
		}()
		return nil
	}, func(_ context.Context) error {
		pool.Close()
		return nil
	}))

	return pool, nil
}
