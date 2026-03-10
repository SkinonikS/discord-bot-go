package tempVoiceChannel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type eventListener struct {
	service Service
	log     *zap.SugaredLogger
}

type EventListenerParams struct {
	fx.In

	Service Service
	Log     *zap.Logger
}

func NewEventListener(p EventListenerParams) *disgoevents.ListenerAdapter {
	el := &eventListener{
		service: p.Service,
		log:     p.Log.Sugar(),
	}

	return &disgoevents.ListenerAdapter{
		OnGuildVoiceStateUpdate: el.GuildVoiceStateUpdate,
	}
}

func (el *eventListener) GuildVoiceStateUpdate(e *disgoevents.GuildVoiceStateUpdate) {
	err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		leftChannelID := snowflake.ID(0)
		if e.OldVoiceState.ChannelID != nil {
			leftChannelID = *e.OldVoiceState.ChannelID
		}

		newChannelID := snowflake.ID(0)
		if e.VoiceState.ChannelID != nil {
			newChannelID = *e.VoiceState.ChannelID
		}

		if leftChannelID == newChannelID {
			return nil
		}

		if leftChannelID != 0 {
			if err := el.service.LeaveChannel(
				ctx,
				e.VoiceState.GuildID,
				leftChannelID,
			); err != nil {
				if errors.Is(err, ErrNotTempChannel) {
					return nil
				}

				el.log.Errorw("failed to handle leave", zap.Error(err))
			}
		}

		if newChannelID != 0 {
			if e.Member.User.Bot {
				return nil
			}

			ownerName := e.Member.User.Username
			if e.Member.User.GlobalName != nil {
				ownerName = *e.Member.User.GlobalName
			}

			if _, err := el.service.JoinChannel(ctx, JoinChannel{
				SetupChannelID: newChannelID,
				GuildID:        e.VoiceState.GuildID,
				OwnerID:        e.Member.User.ID,
				OwnerName:      ownerName,
			}); err != nil {
				if errors.Is(err, ErrNotSetupChannel) {
					return nil
				}

				return fmt.Errorf("failed to handle join: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		el.log.Errorw("failed to handle voice state update", zap.Error(err))
	}
}
