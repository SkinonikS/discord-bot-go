package tempVoiceChannel

import (
	"context"
	"errors"
	"fmt"

	disgocache "github.com/disgoorg/disgo/cache"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrNotSetupChannel      = errors.New("not setup channel")
	ErrSetupChannelNotFound = errors.New("setup channel not found")
	ErrNotTempChannel       = errors.New("not temp channel")
)

type Service interface {
	LeaveChannel(ctx context.Context, params LeaveChannel) error
	JoinChannel(ctx context.Context, params JoinChannel) (disgodiscord.GuildChannel, error)
	CreateSetupChannel(ctx context.Context, params SetupChannel) (*Channel, error)
	DeleteSetupChannel(ctx context.Context, params DeleteSetupChannel) error
}

type serviceImpl struct {
	discordApi       disgorest.Rest
	botCache         disgocache.Caches
	channelRepo      ChannelRepo
	channelStateRepo ChannelStateRepo
	log              *zap.SugaredLogger
}

type ServiceParams struct {
	fx.In

	DiscordApi       disgorest.Rest
	BotCache         disgocache.Caches
	ChannelRepo      ChannelRepo
	ChannelStateRepo ChannelStateRepo
	Log              *zap.Logger
}

func NewService(p ServiceParams) Service {
	return &serviceImpl{
		discordApi:       p.DiscordApi,
		botCache:         p.BotCache,
		channelRepo:      p.ChannelRepo,
		channelStateRepo: p.ChannelStateRepo,
		log:              p.Log.Sugar(),
	}
}

type DeleteSetupChannel struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
}

func (s *serviceImpl) DeleteSetupChannel(ctx context.Context, params DeleteSetupChannel) error {
	setupChannels, err := s.channelRepo.FindByCriteria(ctx, ChannelSearchCriteria{
		GuildID:       params.GuildID,
		RootChannelID: params.ChannelID,
	})
	if err != nil {
		return fmt.Errorf("failed to find temp voice channel setup: %w", err)
	}
	if len(setupChannels) == 0 {
		return ErrSetupChannelNotFound
	}

	_, err = s.channelRepo.DeleteManyByIDs(ctx, lo.Map(setupChannels, func(setupChannel Channel, _ int) uuid.UUID {
		return setupChannel.ID
	}))
	return err
}

type SetupChannel struct {
	GuildID          snowflake.ID
	RootChannelID    snowflake.ID
	ParentCategoryID snowflake.ID
}

func (s *serviceImpl) CreateSetupChannel(ctx context.Context, sc SetupChannel) (*Channel, error) {
	setupChannel := &Channel{
		GuildID:       sc.GuildID,
		RootChannelID: sc.RootChannelID,
		ParentID:      sc.ParentCategoryID,
	}

	if err := s.channelRepo.Save(ctx, setupChannel); err != nil {
		return nil, fmt.Errorf("failed to save temp voice channel setup: %w", err)
	}

	return setupChannel, nil
}

type LeaveChannel struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
}

func (s *serviceImpl) LeaveChannel(ctx context.Context, params LeaveChannel) error {
	channelStates, err := s.channelStateRepo.FindByCriteria(ctx, ChannelStateSearchCriteria{
		GuildID:   params.GuildID,
		ChannelID: params.ChannelID,
	})
	if err != nil {
		return fmt.Errorf("failed to find temp channel state: %w", err)
	}
	if len(channelStates) == 0 {
		return ErrNotTempChannel
	}

	channelState := channelStates[0]
	if s.voiceChannelMembersCount(params.GuildID, params.ChannelID) == 0 {
		return s.channelStateRepo.Transaction(ctx, func(tx ChannelStateRepo) error {
			if err := s.discordApi.DeleteChannel(params.ChannelID, disgorest.WithCtx(ctx)); err != nil {
				return fmt.Errorf("failed to delete voice channel: %w", err)
			}

			if _, err := tx.DeleteManyByIDs(ctx, []uuid.UUID{channelState.ID}); err != nil {
				return fmt.Errorf("failed to delete temp voice channel state: %w", err)
			}

			return nil
		})
	}

	return nil
}

type JoinChannel struct {
	SetupChannelID snowflake.ID
	GuildID        snowflake.ID
	OwnerID        snowflake.ID
	OwnerName      string
}

func (s *serviceImpl) JoinChannel(ctx context.Context, params JoinChannel) (disgodiscord.GuildChannel, error) {
	setupChannel, err := s.channelRepo.FindByCriteria(ctx, ChannelSearchCriteria{
		GuildID:       params.GuildID,
		RootChannelID: params.SetupChannelID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to look up temp voice channel setup: %w", err)
	}
	if len(setupChannel) == 0 {
		return nil, ErrNotSetupChannel
	}

	voiceChannel, err := s.createChannel(ctx, params, setupChannel[0])
	if err != nil {
		return nil, fmt.Errorf("failed to create voice channel: %w", err)
	}

	return voiceChannel, nil
}

func (s *serviceImpl) createChannel(ctx context.Context, params JoinChannel, setupChannel Channel) (disgodiscord.GuildChannel, error) {
	var newVoiceChannel disgodiscord.GuildChannel
	channelName := fmt.Sprintf("🔊 %s's channel", params.OwnerName)
	if err := s.channelStateRepo.Transaction(ctx, func(tx ChannelStateRepo) error {
		var err error
		newVoiceChannel, err = s.discordApi.CreateGuildChannel(params.GuildID, disgodiscord.GuildVoiceChannelCreate{
			Name:     channelName,
			ParentID: setupChannel.ParentID,
			PermissionOverwrites: []disgodiscord.PermissionOverwrite{
				disgodiscord.MemberPermissionOverwrite{
					UserID: params.OwnerID,
					Allow: disgodiscord.PermissionManageChannels.Add(
						disgodiscord.PermissionMoveMembers,
						disgodiscord.PermissionMuteMembers,
						disgodiscord.PermissionDeafenMembers,
					),
				},
			},
		}, disgorest.WithCtx(ctx))
		if err != nil {
			return fmt.Errorf("failed to create voice channel: %w", err)
		}

		if _, err := s.discordApi.UpdateMember(params.GuildID, params.OwnerID, disgodiscord.MemberUpdate{
			ChannelID: new(newVoiceChannel.ID()),
		}, disgorest.WithCtx(ctx)); err != nil {
			return fmt.Errorf("failed to move user to voice channel: %w", err)
		}

		if err := tx.Save(ctx, &ChannelState{
			ChannelID: newVoiceChannel.ID(),
			GuildID:   params.GuildID,
		}); err != nil {
			return fmt.Errorf("failed to save temp voice channel state: %w", err)
		}

		return nil
	}); err != nil {
		_ = s.discordApi.DeleteChannel(newVoiceChannel.ID(), disgorest.WithCtx(ctx))
		return nil, err
	}

	return newVoiceChannel, nil
}

func (s *serviceImpl) voiceChannelMembersCount(guildID, channelID snowflake.ID) uint64 {
	var count uint64
	for vs := range s.botCache.VoiceStates(guildID) {
		if vs.ChannelID == nil {
			continue
		}
		if *vs.ChannelID == channelID {
			count++
		}
	}
	return count
}
