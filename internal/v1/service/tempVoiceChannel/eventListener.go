package tempVoiceChannel

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	discord2 "github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type eventListener struct {
	channelRepo *repo.TempVoiceChannelRepo
	stateRepo   *repo.TempVoiceChannelStateRepo
	log         *zap.SugaredLogger
}

type EventListenerParams struct {
	fx.In
	ChannelRepo *repo.TempVoiceChannelRepo
	StateRepo   *repo.TempVoiceChannelStateRepo
	Log         *zap.Logger
}

func NewEventListener(p EventListenerParams) *disgoevents.ListenerAdapter {
	el := &eventListener{
		channelRepo: p.ChannelRepo,
		stateRepo:   p.StateRepo,
		log:         p.Log.Sugar(),
	}

	return &disgoevents.ListenerAdapter{
		OnGuildVoiceStateUpdate: el.GuildVoiceStateUpdate,
	}
}

func (el *eventListener) GuildVoiceStateUpdate(e *disgoevents.GuildVoiceStateUpdate) {
	err := discord2.ListenWithError(func() error {
		ctx, cancel := discord2.DefaultEventListenContext()
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
			if err := el.handleLeave(ctx, e, leftChannelID); err != nil {
				el.log.Errorw("failed to handle leave", zap.Error(err))
			}
		}

		if newChannelID != 0 {
			if err := el.handleJoin(ctx, e, newChannelID); err != nil {
				return fmt.Errorf("failed to handle join: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		el.log.Errorw("failed to handle voice state update", zap.Error(err))
	}
}

func (el *eventListener) handleLeave(ctx context.Context, e *disgoevents.GuildVoiceStateUpdate, channelID snowflake.ID) error {
	state, err := el.stateRepo.FindByChannelID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to find temp channel state: %w", err)
	}
	if state == nil {
		return nil
	}

	if countChannelMembers(e.Client(), e.VoiceState.GuildID, channelID) == 0 {
		if err := e.Client().Rest.DeleteChannel(channelID, disgorest.WithCtx(ctx)); err != nil {
			return fmt.Errorf("failed to delete voice channel: %w", err)
		}
		if err := el.stateRepo.DeleteByChannelID(ctx, channelID); err != nil {
			return fmt.Errorf("failed to delete temp voice channel state: %w", err)
		}
	}

	return nil
}

func (el *eventListener) handleJoin(ctx context.Context, e *disgoevents.GuildVoiceStateUpdate, channelID snowflake.ID) error {
	state, err := el.stateRepo.FindByChannelID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to find temp channel state: %w", err)
	}
	if state != nil {
		return nil
	}

	channel, err := el.channelRepo.FindByRootChannel(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to look up temp voice channel setup: %w", err)
	}
	if channel == nil {
		return nil
	}

	if _, err := el.createChannel(ctx, e, channel.ParentID); err != nil {
		return fmt.Errorf("failed to create voice channel: %w", err)
	}

	return nil
}

func (el *eventListener) createChannel(ctx context.Context, e *disgoevents.GuildVoiceStateUpdate, parentID snowflake.ID) (*disgodiscord.GuildChannel, error) {
	if e.Member.User.Bot {
		return nil, nil
	}

	channelName := fmt.Sprintf("🔊 %s's channel", *e.Member.User.GlobalName)
	newVoiceChannel, err := e.Client().Rest.CreateGuildChannel(e.VoiceState.GuildID, disgodiscord.GuildVoiceChannelCreate{
		Name: channelName,
		PermissionOverwrites: []disgodiscord.PermissionOverwrite{
			disgodiscord.MemberPermissionOverwrite{
				UserID: e.Member.User.ID,
				Allow: disgodiscord.Permissions.Add(
					disgodiscord.PermissionManageChannels,
					disgodiscord.PermissionMoveMembers,
					disgodiscord.PermissionMuteMembers,
					disgodiscord.PermissionDeafenMembers,
				),
			},
		},
		ParentID: parentID,
	}, disgorest.WithCtx(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create voice channel: %w", err)
	}

	if _, err := e.Client().Rest.UpdateMember(e.VoiceState.GuildID, e.Member.User.ID, disgodiscord.MemberUpdate{
		ChannelID: new(newVoiceChannel.ID()),
	}, disgorest.WithCtx(ctx)); err != nil {
		_ = e.Client().Rest.DeleteChannel(newVoiceChannel.ID(), disgorest.WithCtx(ctx))
		return nil, fmt.Errorf("failed to move user to voice channel: %w", err)
	}

	if err := el.stateRepo.Save(ctx, &model.TempVoiceChannelState{
		ChannelID: newVoiceChannel.ID(),
	}); err != nil {
		_ = e.Client().Rest.DeleteChannel(newVoiceChannel.ID(), disgorest.WithCtx(ctx))
		return nil, fmt.Errorf("failed to save temp voice channel state: %w", err)
	}

	return &newVoiceChannel, nil
}
