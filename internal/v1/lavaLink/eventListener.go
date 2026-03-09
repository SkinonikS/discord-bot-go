package lavaLink

import (
	"context"

	disgobot "github.com/disgoorg/disgo/bot"
	disgoevents "github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"go.uber.org/fx"
)

type eventListener struct {
	client   *disgobot.Client
	lavaLink disgolink.Client
}

type EventListenerParams struct {
	fx.In
	Client   *disgobot.Client
	LavaLink disgolink.Client
}

func NewEventListener(p EventListenerParams) *disgoevents.ListenerAdapter {
	el := &eventListener{
		client:   p.Client,
		lavaLink: p.LavaLink,
	}

	return &disgoevents.ListenerAdapter{
		OnGuildVoiceStateUpdate: el.GuildVoiceStateUpdate,
		OnVoiceServerUpdate:     el.VoiceServerUpdate,
	}
}

func (el *eventListener) GuildVoiceStateUpdate(e *disgoevents.GuildVoiceStateUpdate) {
	if e.VoiceState.UserID != el.client.ApplicationID {
		return
	}

	el.lavaLink.OnVoiceStateUpdate(context.Background(), e.VoiceState.GuildID, e.VoiceState.ChannelID, e.VoiceState.SessionID)
}

func (el *eventListener) VoiceServerUpdate(e *disgoevents.VoiceServerUpdate) {
	el.lavaLink.OnVoiceServerUpdate(context.Background(), e.GuildID, e.Token, *e.Endpoint)
}
