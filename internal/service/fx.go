package sharedservice

import (
	"github.com/SkinonikS/discord-bot-go/internal/service/auto_role"
	"github.com/SkinonikS/discord-bot-go/internal/service/info"
	"github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	"github.com/SkinonikS/discord-bot-go/internal/service/music_player"
	"github.com/SkinonikS/discord-bot-go/internal/service/ping"
	"github.com/SkinonikS/discord-bot-go/internal/service/reaction_role"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository"
	"github.com/SkinonikS/discord-bot-go/internal/service/temp_voice_channel"
	"go.uber.org/fx"
)

const (
	ModuleName = "service"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		repository.NewModule(),
		tempvoicechannel.NewModule(),
		reactionrole.NewModule(),
		autorole.NewModule(),
		interactioncommand.NewModule(),
		musicplayer.NewModule(),
		info.NewModule(),
		ping.NewModule(),
	)
}
