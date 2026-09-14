package migrator

import (
	"database/sql"
	"io/fs"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

func New(dialect goose.Dialect, db *sql.DB, migrationsDir fs.FS, log *zap.SugaredLogger) (*goose.Provider, error) {
	return goose.NewProvider(dialect, db, migrationsDir,
		goose.WithLogger(NewLogger(log)),
	)
}
