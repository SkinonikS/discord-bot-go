package musicplayer

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/lavalink"
	"github.com/disgoorg/disgolink/v3/disgolink"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
	"go.uber.org/fx"
)

type DiscordPlayerEventListener struct {
	service Service
}

type PlayerEventListenerParams struct {
	fx.In

	Service Service
}

func NewDiscordPlayerEventListener(p PlayerEventListenerParams) *lavalink.ListenerAdapter {
	el := &DiscordPlayerEventListener{
		service: p.Service,
	}

	return &lavalink.ListenerAdapter{
		OnTrackStart: el.TrackStart,
		OnQueueEnd:   el.QueueEnd,
	}
}

func (el *DiscordPlayerEventListener) TrackStart(player disgolink.Player, _ *disgolavalink.TrackStartEvent) {
	el.service.StopTimer(player.GuildID())
}

func (el *DiscordPlayerEventListener) QueueEnd(player disgolink.Player, _ *lavaqueue.QueueEndEvent) {
	el.service.StartTimer(StartTimer{
		GuildID: player.GuildID(),
		Track: func() *disgolavalink.Track {
			return player.Track()
		},
	})
}
