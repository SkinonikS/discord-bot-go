package sharedservice

import (
	"github.com/SkinonikS/discord-bot-go/internal/service/auto_role"
	"github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	"github.com/SkinonikS/discord-bot-go/internal/service/music_player"
	"github.com/SkinonikS/discord-bot-go/internal/service/reaction_role"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository"
	"github.com/SkinonikS/discord-bot-go/internal/service/temp_voice_channel"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "service"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		tempvoicechannel.NewModule(),
		reactionrole.NewModule(),
		autorole.NewModule(),
		interactioncommand.NewModule(),
		musicplayer.NewModule(),
		repository.NewModule(),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
