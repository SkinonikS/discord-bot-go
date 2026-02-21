package database

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/migrator"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "database"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(migrator.New, New),
		fx.Provide(
			repo.NewReactionRoleRepo,
			repo.NewTempVoiceChannelRepo,
			repo.NewTempVoiceChannelStateRepo,
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
		fx.Invoke(func(lc fx.Lifecycle, migrator *goose.Provider, log *zap.Logger) {
			lc.Append(fx.StartHook(func(ctx context.Context) error {
				if _, err := migrator.Up(ctx); err != nil {
					return err
				}
				log.Info("database migrations applied")
				return nil
			}))
		}),
	)
}
