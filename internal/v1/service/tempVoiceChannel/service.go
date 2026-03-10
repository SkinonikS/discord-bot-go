package tempVoiceChannel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	disgocache "github.com/disgoorg/disgo/cache"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrNotSetupChannel      = errors.New("not setup channel")
	ErrSetupChannelNotFound = errors.New("setup channel not found")
	ErrNotTempChannel       = errors.New("not temp channel")
)

type Service interface {
	LeaveChannel(ctx context.Context, guildID, channelID snowflake.ID) error
	JoinChannel(ctx context.Context, joinChannel JoinChannel) (disgodiscord.GuildChannel, error)
	CreateSetupChannel(ctx context.Context, sc SetupChannel) (*Channel, error)
	DeleteSetupChannel(ctx context.Context, guildID, channelID snowflake.ID) error
	DeleteStaleChannels(ctx context.Context, shardID uint32) error
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

func (s *serviceImpl) DeleteSetupChannel(
	ctx context.Context,
	guildID, channelID snowflake.ID,
) error {
	channel, err := s.channelRepo.Find(ctx, guildID, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return fmt.Errorf("failed to find temp voice channel setup: %w", err)
	}

	_, err = s.channelRepo.DeleteByID(ctx, channel.ID)
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

func (s *serviceImpl) LeaveChannel(ctx context.Context, guildID, channelID snowflake.ID) error {
	state, err := s.channelStateRepo.Find(ctx, guildID, channelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotTempChannel
		}
		return fmt.Errorf("failed to find temp channel state: %w", err)
	}

	if s.voiceChannelMembersCount(guildID, channelID) == 0 {
		if err := s.channelStateRepo.Transaction(ctx, func(tx ChannelStateRepo) error {
			if err := s.discordApi.DeleteChannel(channelID, disgorest.WithCtx(ctx)); err != nil {
				return fmt.Errorf("failed to delete voice channel: %w", err)
			}

			if _, err := tx.DeleteByID(ctx, state.ID); err != nil {
				return fmt.Errorf("failed to delete temp voice channel state: %w", err)
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

type JoinChannel struct {
	SetupChannelID snowflake.ID
	GuildID        snowflake.ID
	OwnerID        snowflake.ID
	OwnerName      string
}

func (s *serviceImpl) JoinChannel(
	ctx context.Context,
	jc JoinChannel,
) (disgodiscord.GuildChannel, error) {
	channel, err := s.channelRepo.Find(ctx, jc.GuildID, jc.SetupChannelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotSetupChannel
		}
		return nil, fmt.Errorf("failed to look up temp voice channel setup: %w", err)
	}

	voiceChannel, err := s.createChannel(ctx, jc, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to create voice channel: %w", err)
	}

	return voiceChannel, nil
}

func (s *serviceImpl) DeleteStaleChannels(ctx context.Context, shardID uint32) error {
	const batchSize = 100
	if err := s.channelStateRepo.FindInBatchesByShard(
		ctx,
		shardID,
		batchSize,
		func(states []ChannelState, _ int) error {
			for _, state := range states {
				channel, err := s.discordApi.GetChannel(state.ChannelID, disgorest.WithCtx(ctx))
				if err != nil {
					if restErr, ok := errors.AsType[*disgorest.Error](
						err,
					); ok && restErr.Response != nil &&
						restErr.Response.StatusCode == http.StatusNotFound {
						if _, err := s.channelStateRepo.Delete(
							ctx,
							state.GuildID,
							state.ChannelID,
						); err != nil {
							s.log.Warnw(
								"failed to delete orphaned record",
								"channelID",
								state.ChannelID,
								zap.Error(err),
							)
						}
					}
					continue
				}

				if _, ok := channel.(disgodiscord.GuildChannel); !ok {
					continue
				}

				if s.voiceChannelMembersCount(state.GuildID, state.ChannelID) == 0 {
					if err := s.discordApi.DeleteChannel(
						state.ChannelID,
						disgorest.WithCtx(ctx),
					); err != nil {
						s.log.Warnw(
							"failed to delete empty channel",
							"channelID",
							state.ChannelID,
							zap.Error(err),
						)
					}

					if _, err := s.channelStateRepo.Delete(
						ctx,
						state.GuildID,
						state.ChannelID,
					); err != nil {
						s.log.Warnw(
							"failed to delete state record",
							"channelID",
							state.ChannelID,
							zap.Error(err),
						)
					}
				}

				time.Sleep(1 * time.Second)
			}
			return nil
		},
	); err != nil {
		return fmt.Errorf("failed to fetch active channels: %w", err)
	}

	return nil
}

func (s *serviceImpl) createChannel(
	ctx context.Context,
	jc JoinChannel,
	setupChannel *Channel,
) (disgodiscord.GuildChannel, error) {
	var newVoiceChannel disgodiscord.GuildChannel
	channelName := fmt.Sprintf("🔊 %s's channel", jc.OwnerName)
	if err := s.channelStateRepo.Transaction(ctx, func(tx ChannelStateRepo) error {
		var err error
		newVoiceChannel, err = s.discordApi.CreateGuildChannel(
			jc.GuildID,
			disgodiscord.GuildVoiceChannelCreate{
				Name:     channelName,
				ParentID: setupChannel.ParentID,
				PermissionOverwrites: []disgodiscord.PermissionOverwrite{
					disgodiscord.MemberPermissionOverwrite{
						UserID: jc.OwnerID,
						Allow: disgodiscord.PermissionManageChannels.Add(
							disgodiscord.PermissionMoveMembers,
							disgodiscord.PermissionMuteMembers,
							disgodiscord.PermissionDeafenMembers,
						),
					},
				},
			},
			disgorest.WithCtx(ctx),
		)
		if err != nil {
			return fmt.Errorf("failed to create voice channel: %w", err)
		}

		if _, err := s.discordApi.UpdateMember(jc.GuildID, jc.OwnerID, disgodiscord.MemberUpdate{
			ChannelID: new(newVoiceChannel.ID()),
		}, disgorest.WithCtx(ctx)); err != nil {
			return fmt.Errorf("failed to move user to voice channel: %w", err)
		}

		if err := tx.Save(ctx, &ChannelState{
			ChannelID: newVoiceChannel.ID(),
			GuildID:   jc.GuildID,
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
