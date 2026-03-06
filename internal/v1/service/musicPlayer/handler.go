package musicPlayer

import (
	"context"
	"time"

	"github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	"go.uber.org/zap"
)

type eventListener struct {
	manager *Manager
	log     *zap.SugaredLogger
}

func NewEventListener(manager *Manager, log *zap.Logger) *disgoevents.ListenerAdapter {
	el := &eventListener{
		manager: manager,
		log:     log.Sugar(),
	}

	return &disgoevents.ListenerAdapter{
		OnGuildVoiceStateUpdate: el.GuildVoiceStateUpdate,
	}
}

func (el *eventListener) GuildVoiceStateUpdate(e *disgoevents.GuildVoiceStateUpdate) {
	err := discord.ListenWithError(func() error {
		selfUserID := e.Client().ID()
		if e.VoiceState.UserID != selfUserID {
			return nil
		}

		if e.VoiceState.ChannelID == nil {
			return el.manager.Disconnect(e.VoiceState.GuildID)
		}

		if !el.manager.HasGuildState(e.VoiceState.GuildID) {
			conn := e.Client().VoiceManager.GetConn(e.VoiceState.GuildID)
			if conn != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				conn.Close(ctx)
			}
		}

		return nil
	})
	if err != nil {
		el.log.Errorw("failed to handle voice state update", zap.Error(err))
	}
}
