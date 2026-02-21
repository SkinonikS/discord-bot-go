package tempVoiceChannel

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/SkinonikS/discord-bot-go/internal/v1/util"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Handler struct {
	channelRepo *repo.TempVoiceChannelRepo
	stateRepo   *repo.TempVoiceChannelStateRepo
	log         *zap.SugaredLogger
}

type HandlerParams struct {
	fx.In
	ChannelRepo *repo.TempVoiceChannelRepo
	StateRepo   *repo.TempVoiceChannelStateRepo
	Log         *zap.Logger
}

func NewHandler(p HandlerParams) *Handler {
	return &Handler{
		channelRepo: p.ChannelRepo,
		stateRepo:   p.StateRepo,
		log:         p.Log.Sugar(),
	}
}

func (h *Handler) Handle(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	err := util.Safe(func() error {
		ctx, cancel := discord.DefaultHandlerContext()
		defer cancel()

		if e.Member.User.Bot {
			return nil
		}

		leftChannelID := ""
		if e.BeforeUpdate != nil {
			leftChannelID = e.BeforeUpdate.ChannelID
		}

		if leftChannelID == e.ChannelID {
			return nil
		}

		if leftChannelID != "" {
			if err := h.handleLeave(ctx, s, leftChannelID); err != nil {
				h.log.Errorw("failed to handle leave", zap.Error(err))
			}
		}

		if e.ChannelID != "" {
			if err := h.handleJoin(ctx, s, e); err != nil {
				return fmt.Errorf("failed to handle join: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		h.log.Errorw("failed to handle voice state update", zap.Error(err))
	}
}

func (h *Handler) handleLeave(ctx context.Context, s *discordgo.Session, channelID string) error {
	state, err := h.stateRepo.FindByChannelID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to find temp channel state: %w", err)
	}
	if state == nil {
		return nil
	}

	if err := h.stateRepo.DecrementMemberCount(ctx, channelID); err != nil {
		return fmt.Errorf("failed to decrement member count: %w", err)
	}

	newState, err := h.stateRepo.FindByChannelID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to decrement member count: %w", err)
	}
	if newState == nil {
		return nil
	}

	if newState.MemberCount <= 0 {
		if _, err := s.ChannelDelete(channelID, discordgo.WithContext(ctx)); err != nil {
			return fmt.Errorf("failed to delete voice channel: %w", err)
		}
		if err := h.stateRepo.DeleteByChannelID(ctx, channelID); err != nil {
			return fmt.Errorf("failed to delete temp voice channel state: %w", err)
		}
	}

	return nil
}

func (h *Handler) handleJoin(ctx context.Context, s *discordgo.Session, e *discordgo.VoiceStateUpdate) error {
	state, err := h.stateRepo.FindByChannelID(ctx, e.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find temp channel state: %w", err)
	}
	if state != nil {
		if err := h.stateRepo.IncrementMemberCount(ctx, e.ChannelID); err != nil {
			return fmt.Errorf("failed to increment member count: %w", err)
		}
		return nil
	}

	channel, err := h.channelRepo.FindByRootChannel(ctx, e.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to look up temp voice channel setup: %w", err)
	}
	if channel == nil {
		return nil
	}

	if _, err := h.createChannel(ctx, s, e, channel.ParentID); err != nil {
		return fmt.Errorf("failed to create voice channel: %w", err)
	}

	return nil
}

func (h *Handler) createChannel(ctx context.Context, s *discordgo.Session, e *discordgo.VoiceStateUpdate, parentID string) (*discordgo.Channel, error) {
	member, err := s.GuildMember(e.GuildID, e.UserID, discordgo.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}
	if member.User.Bot {
		return nil, nil
	}

	channelName := fmt.Sprintf("🔊 Комната %s", member.User.GlobalName)
	newChannel, err := s.GuildChannelCreateComplex(e.GuildID, discordgo.GuildChannelCreateData{
		Name:     channelName,
		Type:     discordgo.ChannelTypeGuildVoice,
		ParentID: parentID,
	}, discordgo.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create voice channel: %w", err)
	}

	if err := s.ChannelPermissionSet(
		newChannel.ID,
		e.UserID,
		discordgo.PermissionOverwriteTypeMember,
		discordgo.PermissionManageChannels|
			discordgo.PermissionVoiceMoveMembers|
			discordgo.PermissionVoiceMuteMembers|
			discordgo.PermissionVoiceDeafenMembers,
		0,
		discordgo.WithContext(ctx),
	); err != nil {
		go s.ChannelDelete(newChannel.ID)
		return nil, fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := s.GuildMemberMove(e.GuildID, e.UserID, &newChannel.ID, discordgo.WithContext(ctx)); err != nil {
		go s.ChannelDelete(newChannel.ID)
		return nil, fmt.Errorf("failed to move user to voice channel: %w", err)
	}

	if err := h.stateRepo.Save(ctx, &model.TempVoiceChannelState{
		ChannelID:   newChannel.ID,
		MemberCount: 1,
	}); err != nil {
		go s.ChannelDelete(newChannel.ID)
		return nil, fmt.Errorf("failed to save temp voice channel state: %w", err)
	}

	return newChannel, nil
}
