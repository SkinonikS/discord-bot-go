package pprof

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"

	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "pprof"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig),
		fx.Invoke(func(lc fx.Lifecycle, cfg *Config, appCfg *config.Config, log *zap.Logger) error {
			if !appCfg.Debug {
				log.Info("app is not in debug mode, pprof server is disabled")
				return nil
			}

			srv := &http.Server{Addr: cfg.Addr}
			lc.Append(fx.StartStopHook(
				func(context.Context) error {
					log.Info("starting pprof server", zap.String("addr", cfg.Addr))
					go func() {
						if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
							log.Error("pprof server error", zap.Error(err))
						}
					}()
					return nil
				},
				func(ctx context.Context) error {
					return srv.Shutdown(ctx)
				},
			))

			return nil
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
