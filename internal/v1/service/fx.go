package service

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayer"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayerSource"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/reactionRole"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/tempVoiceChannel"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/youtubeMusicPlayerSource"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "service"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		tempVoiceChannel.NewModule(),
		reactionRole.NewModule(),
		interactionCommand.NewModule(),
		youtubeMusicPlayerSource.NewModule(),
		musicPlayerSource.NewModule(),
		musicPlayer.NewModule(),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
