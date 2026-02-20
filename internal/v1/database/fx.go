package database

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	ModuleName = "database"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(New),
		fx.Provide(
			repo.NewTempVoiceChannelSetupRepo,
			repo.NewReactionRoleRepo,
			repo.NewTempVoiceChannelStateRepo,
		),
		fx.Invoke(func(*gorm.DB) {}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
