package lavaLink

import (
	"github.com/disgoorg/disgolink/v3/disgolink"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
)

type ListenerAdapter struct {
	OnTrackStart func(disgolink.Player, *disgolavalink.TrackStartEvent)
	OnQueueEnd   func(disgolink.Player, *lavaqueue.QueueEndEvent)
}

func (la *ListenerAdapter) OnEvent(player disgolink.Player, event disgolavalink.Message) {
	switch e := event.(type) {
	case *disgolavalink.TrackStartEvent:
		if la.OnTrackStart != nil {
			la.OnTrackStart(player, e)
		}
	case *lavaqueue.QueueEndEvent:
		if la.OnQueueEnd != nil {
			la.OnQueueEnd(player, e)
		}
	}
}
