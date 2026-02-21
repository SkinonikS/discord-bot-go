package migrator

import (
	"os"

	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Params struct {
	fx.In
	Log  *zap.Logger
	DB   *gorm.DB
	Path *foundation.Path
}

func New(p Params) (*goose.Provider, error) {
	rawDB, err := p.DB.DB()
	if err != nil {
		return nil, err
	}

	migrationsPath := os.DirFS(p.Path.MigrationsPath())
	provider, err := goose.NewProvider(goose.DialectSQLite3, rawDB, migrationsPath,
		goose.WithLogger(NewLogger(p.Log)),
	)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
