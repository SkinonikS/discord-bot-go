package database

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "database"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig, New),
		fx.Provide(
			repo.NewReactionRoleRepo,
			repo.NewTempVoiceChannelRepo,
			repo.NewTempVoiceChannelStateRepo,
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
