package postgres

import (
	"fmt"
	"os"

	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/infra/migrator"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MigratorParams struct {
	fx.In

	Pool *pgxpool.Pool
	Path *foundation.Path
	Log  *zap.Logger
}

func newMigratorProvider(p MigratorParams) (migrator.Provider, error) {
	migrationsDir := os.DirFS(p.Path.MigrationsPath(ModuleName))

	provider, err := migrator.New(goose.DialectPostgres, stdlib.OpenDBFromPool(p.Pool), migrationsDir, p.Log.Sugar())
	if err != nil {
		return migrator.Provider{}, fmt.Errorf("failed to create postgres migrator: %w", err)
	}

	return migrator.Provider{
		Name:     ModuleName,
		Provider: provider,
	}, nil
}
