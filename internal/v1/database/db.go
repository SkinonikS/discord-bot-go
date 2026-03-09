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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

type Params struct {
	fx.In
	Lc     fx.Lifecycle
	Log    *zap.Logger
	Path   *foundation.Path
	Config *Config
}

type Result struct {
	fx.Out
	DB       *gorm.DB
	Migrator *goose.Provider
}

func New(p Params) (Result, error) {
	dialector := postgres.Open(fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		p.Config.Host,
		p.Config.Port,
		p.Config.User,
		p.Config.DB,
		p.Config.Password,
		p.Config.SSLMode,
	))

	zapGormLog := zapgorm2.New(p.Log)
	zapGormLog.IgnoreRecordNotFoundError = true
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:               zapGormLog.LogMode(gormlogger.Info),
		DisableAutomaticPing: true,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to connect database: %w", err)
	}

	dbMigrator, err := migrator.New(migrator.Params{
		Log:           p.Log.Sugar(),
		DB:            db,
		MigrationsDir: os.DirFS(p.Path.MigrationsPath()),
	})

	p.Lc.Append(fx.StartStopHook(
		func(ctx context.Context) error {
			go func() {
				rawDB, err := db.DB()
				if err != nil {
					p.Log.Error("failed to get database", zap.Error(err))
					return
				}
				if err := rawDB.PingContext(ctx); err != nil {
					p.Log.Error("failed to ping database", zap.Error(err))
					return
				}
				p.Log.Info("database connection established")

				if _, err := dbMigrator.Up(ctx); err != nil {
					p.Log.Error("failed to apply database migrations", zap.Error(err))
					return
				}
				p.Log.Info("database migrations applied")
			}()
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

			p.Log.Info("database connection closed")
			return nil
		},
	))

	return Result{
		DB:       db,
		Migrator: dbMigrator,
	}, nil
}
