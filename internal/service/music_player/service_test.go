package musicplayer_test

import (
	"context"
	"testing"
	"time"

	discordmock "github.com/SkinonikS/discord-bot-go/internal/infra/discord/mock"
	musicplayer "github.com/SkinonikS/discord-bot-go/internal/service/music_player"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newService(t *testing.T, idleTimeout time.Duration, client *disgobot.Client) musicplayer.Service {
	t.Helper()

	if client == nil {
		client = &disgobot.Client{}
	}

	return musicplayer.NewService(musicplayer.ServiceParams{
		Config:    &musicplayer.Config{IdleTimeout: idleTimeout},
		BotClient: client,
	})
}

func TestService_StopTimer(t *testing.T) {
	t.Run("returns false when nothing is scheduled for the guild", func(t *testing.T) {
		svc := newService(t, time.Minute, nil)

		assert.False(t, svc.StopTimer(snowflake.ID(1)))
	})

	t.Run("stops and forgets a scheduled timer", func(t *testing.T) {
		svc := newService(t, time.Minute, nil)
		guildID := snowflake.ID(1)

		svc.StartTimer(musicplayer.StartTimer{GuildID: guildID, Track: func() *disgolavalink.Track { return &disgolavalink.Track{} }})

		assert.True(t, svc.StopTimer(guildID))
		assert.False(t, svc.StopTimer(guildID))
	})
}

func TestService_StartTimer(t *testing.T) {
	t.Run("replaces an existing timer for the same guild instead of stacking", func(t *testing.T) {
		svc := newService(t, time.Minute, nil)
		guildID := snowflake.ID(1)

		svc.StartTimer(musicplayer.StartTimer{GuildID: guildID, Track: func() *disgolavalink.Track { return &disgolavalink.Track{} }})
		svc.StartTimer(musicplayer.StartTimer{GuildID: guildID, Track: func() *disgolavalink.Track { return &disgolavalink.Track{} }})

		assert.True(t, svc.StopTimer(guildID))
		assert.False(t, svc.StopTimer(guildID))
	})

	t.Run("disconnects from voice once idle if nothing is playing", func(t *testing.T) {
		guildID := snowflake.ID(7)
		gw := discordmock.NewMockGateway(t)
		done := make(chan struct{})
		gw.EXPECT().Send(mock.Anything, gateway.OpcodeVoiceStateUpdate, mock.MatchedBy(func(d gateway.MessageDataVoiceStateUpdate) bool {
			return d.GuildID == guildID && d.ChannelID == nil
		})).Run(func(context.Context, gateway.Opcode, gateway.MessageData) {
			close(done)
		}).Return(nil)

		svc := newService(t, 10*time.Millisecond, &disgobot.Client{Gateway: gw})
		svc.StartTimer(musicplayer.StartTimer{GuildID: guildID, Track: func() *disgolavalink.Track { return nil }})

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for the idle disconnect")
		}
	})

	t.Run("stays connected while a track is still playing when the timer fires", func(t *testing.T) {
		guildID := snowflake.ID(8)
		gw := discordmock.NewMockGateway(t)

		svc := newService(t, 10*time.Millisecond, &disgobot.Client{Gateway: gw})
		svc.StartTimer(musicplayer.StartTimer{GuildID: guildID, Track: func() *disgolavalink.Track { return &disgolavalink.Track{} }})

		time.Sleep(50 * time.Millisecond)
		gw.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
	})
}
