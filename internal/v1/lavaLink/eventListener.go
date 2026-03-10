package lavaLink

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"go.uber.org/fx"
)

type eventListener struct {
	discordConfig *discord.Config
	lavaLink      disgolink.Client
}

type EventListenerParams struct {
	fx.In

	DiscordConfig *discord.Config
	LavaLink      disgolink.Client
}

func NewEventListener(p EventListenerParams) *disgoevents.ListenerAdapter {
	el := &eventListener{
		discordConfig: p.DiscordConfig,
		lavaLink:      p.LavaLink,
	}

	return &disgoevents.ListenerAdapter{
		OnGuildVoiceStateUpdate: el.GuildVoiceStateUpdate,
		OnVoiceServerUpdate:     el.VoiceServerUpdate,
	}
}

func (el *eventListener) GuildVoiceStateUpdate(e *disgoevents.GuildVoiceStateUpdate) {
	if e.VoiceState.UserID != el.discordConfig.AppID {
		return
	}

	el.lavaLink.OnVoiceStateUpdate(
		context.Background(),
		e.VoiceState.GuildID,
		e.VoiceState.ChannelID,
		e.VoiceState.SessionID,
	)
}

func (el *eventListener) VoiceServerUpdate(e *disgoevents.VoiceServerUpdate) {
	el.lavaLink.OnVoiceServerUpdate(context.Background(), e.GuildID, e.Token, *e.Endpoint)
}
