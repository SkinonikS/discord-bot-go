package musicPlayer

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/lavaLink"
	"github.com/disgoorg/disgolink/v3/disgolink"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
	"go.uber.org/fx"
)

type PlayerEventListener struct {
	service Service
}

type PlayerEventListenerParams struct {
	fx.In

	Service Service
}

func NewPlayerEventListener(p PlayerEventListenerParams) *lavaLink.ListenerAdapter {
	el := &PlayerEventListener{
		service: p.Service,
	}

	return &lavaLink.ListenerAdapter{
		OnTrackStart: el.TrackStart,
		OnQueueEnd:   el.QueueEnd,
	}
}

func (el *PlayerEventListener) TrackStart(
	player disgolink.Player,
	_ *disgolavalink.TrackStartEvent,
) {
	el.service.StopTimer(player.GuildID())
}

func (el *PlayerEventListener) QueueEnd(player disgolink.Player, _ *lavaqueue.QueueEndEvent) {
	el.service.StartTimer(StartTimer{
		GuildID: player.GuildID(),
		Track: func() *disgolavalink.Track {
			return player.Track()
		},
	})
}
