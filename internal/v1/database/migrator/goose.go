package migrator

import (
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Params struct {
	Log           *zap.SugaredLogger
	DB            *gorm.DB
	MigrationsDir fs.FS
}

func New(p Params) (*goose.Provider, error) {
	rawDB, err := p.DB.DB()
	if err != nil {
		return nil, err
	}

	var dialect goose.Dialect
	switch p.DB.Dialector.Name() {
	case "postgres":
		dialect = goose.DialectPostgres
	case "sqlite":
		dialect = goose.DialectSQLite3
	default:
		return nil, fmt.Errorf("unsupported database dialect: %s", p.DB.Dialector.Name())
	}

	provider, err := goose.NewProvider(dialect, rawDB, p.MigrationsDir,
		goose.WithLogger(NewLogger(p.Log)),
	)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
