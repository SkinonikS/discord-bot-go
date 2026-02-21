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
	setupRepo *repo.TempVoiceChannelSetupRepo
	stateRepo *repo.TempVoiceChannelStateRepo
	log       *zap.SugaredLogger
}

type HandlerParams struct {
	fx.In
	SetupRepo *repo.TempVoiceChannelSetupRepo
	StateRepo *repo.TempVoiceChannelStateRepo
	Log       *zap.Logger
}

func NewHandler(p HandlerParams) *Handler {
	return &Handler{
		setupRepo: p.SetupRepo,
		stateRepo: p.StateRepo,
		log:       p.Log.Sugar(),
	}
}

func (h *Handler) Handle(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	err := util.Safe(func() error {
		ctx, cancel := discord.DefaultHandlerContext()
		defer cancel()

		if e.Member.User.Bot {
			return nil
		}

		if e.ChannelID == "" {
			if err := h.deleteChannel(ctx, s, e); err != nil {
				return fmt.Errorf("failed to delete voice channel: %w", err)
			}
			return nil
		}

		setup, err := h.setupRepo.FindByRootChannel(ctx, e.GuildID, e.ChannelID)
		if err != nil {
			return fmt.Errorf("failed to look up temp voice channel setup: %w", err)
		}
		if setup == nil {
			return nil
		}

		if _, err := h.createChannel(ctx, s, e, setup.ParentID); err != nil {
			return fmt.Errorf("failed to create voice channel: %w", err)
		}

		return nil
	})
	if err != nil {
		h.log.Errorw("failed to handle voice state update", zap.Error(err))
	}
}

func (h *Handler) createChannel(ctx context.Context, s *discordgo.Session, e *discordgo.VoiceStateUpdate, parentID string) (*discordgo.Channel, error) {
	member, err := s.GuildMember(e.GuildID, e.UserID, discordgo.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}
	if member.User.Bot {
		return nil, nil
	}

	channelName := fmt.Sprintf("🔊 Комната %s", member.User.Username)
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
		GuildID:   e.GuildID,
		UserID:    e.UserID,
		ChannelID: newChannel.ID,
	}); err != nil {
		go s.ChannelDelete(newChannel.ID)
		h.log.Errorw("failed to save temp voice channel state", zap.Error(err))
	}

	return newChannel, nil
}

func (h *Handler) deleteChannel(ctx context.Context, s *discordgo.Session, e *discordgo.VoiceStateUpdate) error {
	states, err := h.stateRepo.FindAllByGuild(ctx, e.GuildID)
	if err != nil {
		return fmt.Errorf("failed to find temp channel states: %w", err)
	}
	if len(states) == 0 {
		return nil
	}

	guild, err := s.Guild(e.GuildID)
	if err != nil {
		return fmt.Errorf("failed to get guild: %w", err)
	}

	for _, state := range states {
		memberCount := 0
		for _, voiceState := range guild.VoiceStates {
			if voiceState.ChannelID == state.ChannelID {
				memberCount++
			}
		}

		if memberCount == 0 {
			if _, err := s.ChannelDelete(state.ChannelID, discordgo.WithContext(ctx)); err != nil {
				return fmt.Errorf("failed to delete voice channel: %w", err)
			}
			if err := h.stateRepo.DeleteByChannelID(ctx, state.ChannelID); err != nil {
				return fmt.Errorf("failed to delete temp voice channel state: %w", err)
			}
		}
	}

	return nil
}
