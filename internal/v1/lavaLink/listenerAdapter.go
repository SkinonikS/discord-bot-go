package lavaLink

import (
	"github.com/disgoorg/disgolink/v3/disgolink"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
)

type ListenerAdapter struct {
	OnPlayerResume    func(disgolink.Player, *disgolavalink.PlayerResumeEvent)
	OnPlayerPause     func(disgolink.Player, *disgolavalink.PlayerPauseEvent)
	OnWebSocketClosed func(disgolink.Player, *disgolavalink.WebSocketClosedEvent)
	OnTrackException  func(disgolink.Player, *disgolavalink.TrackExceptionEvent)
	OnUnknownEvent    func(disgolink.Player, *disgolavalink.UnknownEvent)
	OnTrackStuck      func(disgolink.Player, *disgolavalink.TrackStuckEvent)
	OnTrackStart      func(disgolink.Player, *disgolavalink.TrackStartEvent)
	OnTrackEnd        func(disgolink.Player, *disgolavalink.TrackEndEvent)
	OnQueueEnd        func(disgolink.Player, *lavaqueue.QueueEndEvent)
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
	case *disgolavalink.WebSocketClosedEvent:
		if la.OnWebSocketClosed != nil {
			la.OnWebSocketClosed(player, e)
		}
	case *disgolavalink.PlayerPauseEvent:
		if la.OnPlayerPause != nil {
			la.OnPlayerPause(player, e)
		}
	case *disgolavalink.PlayerResumeEvent:
		if la.OnPlayerResume != nil {
			la.OnPlayerResume(player, e)
		}
	case *disgolavalink.TrackEndEvent:
		if la.OnTrackEnd != nil {
			la.OnTrackEnd(player, e)
		}
	case *disgolavalink.TrackStuckEvent:
		if la.OnTrackStuck != nil {
			la.OnTrackStuck(player, e)
		}
	case *disgolavalink.TrackExceptionEvent:
		if la.OnTrackException != nil {
			la.OnTrackException(player, e)
		}
	case *disgolavalink.UnknownEvent:
		if la.OnUnknownEvent != nil {
			la.OnUnknownEvent(player, e)
		}
	}
}
