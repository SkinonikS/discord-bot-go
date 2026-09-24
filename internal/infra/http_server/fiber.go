package httpserver

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/SkinonikS/discord-bot-go/internal/infra/config"
	fiberzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
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

func New(p Params) (*fiber.App, error) {
	app := fiber.New(fiber.Config{
		AppName: p.AppConfig.Name,
	})
	app.Use(requestid.New())
	app.Use(recoverer.New(recoverer.Config{
		EnableStackTrace: true,
	}))
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: p.Log,
	}))

	for _, handler := range p.Handlers {
		if err := handler.Register(app); err != nil {
			return nil, fmt.Errorf("failed to register handler: %w", err)
		}
	}

	if p.AppConfig.Debug {
		p.Log.Warn("app in debug mode, pprof enabled")
		app.Use(pprof.New())
	}

	p.Lc.Append(fx.StartStopHook(
		func(context.Context) error {
			//nolint:contextcheck,gosec
			go func() {
				addr := net.JoinHostPort(p.Config.Host, strconv.Itoa(p.Config.Port))

				p.Log.Info("starting http server", zap.String("addr", addr))
				if err := app.Listen(addr, fiber.ListenConfig{
					DisableStartupMessage: true,
				}); err != nil {
					p.Log.Error("failed to run http server", zap.Error(err))
				}
			}()
			return nil
		},
		func(ctx context.Context) error {
			p.Log.Info("shutting down http server")
			return app.ShutdownWithContext(ctx)
		},
	))

	return app, nil
}
