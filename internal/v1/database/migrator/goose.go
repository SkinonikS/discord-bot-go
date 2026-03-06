package migrator

import (
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

	provider, err := goose.NewProvider(goose.DialectSQLite3, rawDB, p.MigrationsDir,
		goose.WithLogger(NewLogger(p.Log)),
	)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
