package database

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

func models() []any {
	return []any{
		&model.TempVoiceChannelSetup{},
		&model.ReactionRole{},
		&model.TempVoiceChannelState{},
	}
}

type Params struct {
	fx.In
	Log  *zap.Logger
	Lc   fx.Lifecycle
	Path *foundation.Path
}

func New(p Params) (*gorm.DB, error) {
	log := p.Log.Sugar()

	zapgormlog := zapgorm2.New(p.Log)
	zapgormlog.IgnoreRecordNotFoundError = true
	db, err := gorm.Open(sqlite.Open(p.Path.StoragePath("db.sqlite")), &gorm.Config{
		Logger: zapgormlog.LogMode(gormlogger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(models()...); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

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

	return db, nil
}
