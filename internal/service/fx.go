package service

import (
	"github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	"github.com/SkinonikS/discord-bot-go/internal/service/music_player"
	"github.com/SkinonikS/discord-bot-go/internal/service/reaction_role"
	"github.com/SkinonikS/discord-bot-go/internal/service/temp_voice_channel"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "service"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		temp_voice_channel.NewModule(),
		reaction_role.NewModule(),
		interaction_command.NewModule(),
		music_player.NewModule(),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
