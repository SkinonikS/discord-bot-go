package database

import (
	"context"
	"fmt"
	"os"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/migrator"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

type Params struct {
	fx.In
	Lc   fx.Lifecycle
	Log  *zap.Logger
	Path *foundation.Path
}

type Result struct {
	fx.Out
	DB       *gorm.DB
	Migrator *goose.Provider
}

func New(p Params) (Result, error) {
	log := p.Log.Sugar()

	zapgormlog := zapgorm2.New(p.Log)
	zapgormlog.IgnoreRecordNotFoundError = true
	db, err := gorm.Open(sqlite.Open(p.Path.StoragePath("db.sqlite")), &gorm.Config{
		Logger: zapgormlog.LogMode(gormlogger.Info),
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to connect database: %w", err)
	}

	dbMigrator, err := migrator.New(migrator.Params{
		Log:           log,
		DB:            db,
		MigrationsDir: os.DirFS(p.Path.MigrationsPath()),
	})

	p.Lc.Append(fx.StartStopHook(
		func(ctx context.Context) error {
			rawDB, err := db.DB()
			if err != nil {
				return fmt.Errorf("failed to get database: %w", err)
			}
			if err := rawDB.PingContext(ctx); err != nil {
				return fmt.Errorf("failed to ping database: %w", err)
			}
			log.Info("database connection established")

			if _, err := dbMigrator.Up(ctx); err != nil {
				return err
			}
			log.Info("database migrations applied")
			return nil
		},
		func(context.Context) error {
			rawDB, err := db.DB()
			if err != nil {
				return fmt.Errorf("failed to get database: %w", err)
			}
			if err := rawDB.Close(); err != nil {
				return fmt.Errorf("failed to close database: %w", err)
			}

			log.Info("database connection closed")
			return nil
		},
	))

	return Result{
		DB:       db,
		Migrator: dbMigrator,
	}, nil
}
