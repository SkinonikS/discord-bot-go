package httpServer

import (
	"context"
	"fmt"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"github.com/gin-contrib/graceful"
	"github.com/gin-contrib/pprof"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	Config    *Config
	Lc        fx.Lifecycle
	Log       *zap.Logger
	AppConfig *config.Config
	Handlers  []Handler `group:"http_handlers"`
}

//nolint:gochecknoinits
func init() {
	gin.SetMode(gin.ReleaseMode)
}

func New(p Params) (*graceful.Graceful, error) {
	engine := gin.New()
	engine.Use(ginzap.Ginzap(p.Log, time.DateTime, true))
	engine.Use(ginzap.RecoveryWithZap(p.Log, true))

	for _, handler := range p.Handlers {
		if err := handler.Register(engine); err != nil {
			return nil, fmt.Errorf("failed to register handler: %w", err)
		}
	}

	if p.AppConfig.Debug {
		p.Log.Warn("app in debug mode, pprof enabled")
		pprof.Register(engine)
	}

	router, err := graceful.New(engine,
		graceful.WithAddr(p.Config.Addr),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	p.Lc.Append(fx.StartStopHook(
		func(context.Context) error {
			//nolint:contextcheck,gosec
			go func() {
				p.Log.Info("starting http server")
				if err := router.RunWithContext(context.Background()); err != nil {
					p.Log.Panic("failed to run http server", zap.Error(err))
				}
			}()
			return nil
		},
		func(ctx context.Context) error {
			p.Log.Info("shutting down http server")
			return router.Shutdown(ctx)
		},
	))

	return router, nil
}
