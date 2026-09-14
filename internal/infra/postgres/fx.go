package postgres

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/migrator"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "postgres"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(newConfig, New),
		fx.Provide(
			migrator.AsProvider(newMigratorProvider),
		),
		fx.Invoke(func(*pgxpool.Pool) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
