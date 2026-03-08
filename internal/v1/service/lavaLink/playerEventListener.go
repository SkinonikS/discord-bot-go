package lavaLink

import (
	"context"
	"sync"
	"time"

	"github.com/SkinonikS/discord-bot-go/pkg/v1/lavaLink"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgolink/v3/disgolink"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

type PlayerEventListener struct {
	discordClient *disgobot.Client
	idleTimers    map[snowflake.ID]*time.Timer
	idleTimersMu  sync.Mutex
	config        *Config
}

type PlayerEventListenerParams struct {
	fx.In
	Config        *Config
	DiscordClient *disgobot.Client
}

func NewLavaLinkEventListener(p PlayerEventListenerParams) *lavaLink.ListenerAdapter {
	el := &PlayerEventListener{
		discordClient: p.DiscordClient,
		config:        p.Config,
		idleTimers:    make(map[snowflake.ID]*time.Timer),
	}

	return &lavaLink.ListenerAdapter{
		OnTrackStart: el.TrackStart,
		OnQueueEnd:   el.QueueEnd,
	}
}

func (el *PlayerEventListener) TrackStart(player disgolink.Player, _ *disgolavalink.TrackStartEvent) {
	el.idleTimersMu.Lock()
	defer el.idleTimersMu.Unlock()
	if t, ok := el.idleTimers[player.GuildID()]; ok {
		t.Stop()
		delete(el.idleTimers, player.GuildID())
	}
}

func (el *PlayerEventListener) QueueEnd(player disgolink.Player, e *lavaqueue.QueueEndEvent) {
	el.idleTimersMu.Lock()
	defer el.idleTimersMu.Unlock()

	if t, ok := el.idleTimers[player.GuildID()]; ok {
		t.Stop()
	}

	el.idleTimers[player.GuildID()] = time.AfterFunc(el.config.IdleTimeout, func() {
		el.idleTimersMu.Lock()
		delete(el.idleTimers, player.GuildID())
		el.idleTimersMu.Unlock()

		if player.Track() != nil {
			return
		}

		ctx := context.Background()
		_ = lavaqueue.ClearQueue(ctx, player.Node(), player.GuildID())
		_ = player.Destroy(ctx)
		_ = el.discordClient.UpdateVoiceState(ctx, player.GuildID(), nil, false, false)
	})
}
