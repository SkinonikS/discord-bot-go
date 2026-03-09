package database

import (
	"context"
	"fmt"
	"os"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/migrator"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/gookit/goutil/fsutil"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
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
	var dialector gorm.Dialector
	switch p.Config.Driver {
	case "sqlite":
		if p.Config.Sqlite == nil {
			return Result{}, fmt.Errorf("sqlite config is required")
		}

		var path string
		if fsutil.IsAbsPath(p.Config.Sqlite.Path) {
			path = p.Config.Sqlite.Path
		} else {
			path = p.Path.Path(p.Config.Sqlite.Path)
		}
		dialector = sqlite.Open(path)
	case "postgres":
		if p.Config.Postgres == nil {
			return Result{}, fmt.Errorf("postgres config is required")
		}

		dialector = postgres.Open(fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			p.Config.Postgres.Host,
			p.Config.Postgres.Port,
			p.Config.Postgres.User,
			p.Config.Postgres.DB,
			p.Config.Postgres.Password,
			p.Config.Postgres.SSLMode,
		))
	default:
		return Result{}, fmt.Errorf("unsupported database driver: %s", p.Config.Driver)
	}

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
				p.Log.Info("database connection established", zap.String("driver", p.Config.Driver))

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
