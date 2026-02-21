package reactionRole

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
	repo *repo.ReactionRoleRepo
	log  *zap.SugaredLogger
}

type HandlerParams struct {
	fx.In
	Repo *repo.ReactionRoleRepo
	Log  *zap.Logger
}

func NewHandler(p HandlerParams) *Handler {
	return &Handler{
		repo: p.Repo,
		log:  p.Log.Sugar(),
	}
}

func (h *Handler) HandleAdd(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	err := util.Safe(func() error {
		ctx, cancel := discord.DefaultHandlerContext()
		defer cancel()

		if e.UserID == "" || e.MessageID == "" || e.Member.User.Bot {
			return nil
		}

		reaction, err := h.findEmoji(ctx, e.Emoji, e.MessageID)
		if err != nil {
			return err
		}
		if reaction == nil {
			return nil
		}

		if err := s.GuildMemberRoleAdd(e.GuildID, e.UserID, reaction.RoleID, discordgo.WithContext(ctx)); err != nil {
			return fmt.Errorf("failed to add reaction role: %w", err)
		}

		return nil
	})
	if err != nil {
		h.log.Errorw("failed to handle reaction role", zap.Error(err))
	}
}

func (h *Handler) HandleRemove(s *discordgo.Session, e *discordgo.MessageReactionRemove) {
	err := util.Safe(func() error {
		ctx, cancel := discord.DefaultHandlerContext()
		defer cancel()

		if e.UserID == "" || e.MessageID == "" {
			return nil
		}

		user, err := s.User(e.UserID, discordgo.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		if user.Bot {
			return nil
		}

		reaction, err := h.findEmoji(ctx, e.Emoji, e.MessageID)
		if err != nil {
			return err
		}
		if reaction == nil {
			return nil
		}

		if err := s.GuildMemberRoleRemove(e.GuildID, e.UserID, reaction.RoleID, discordgo.WithContext(ctx)); err != nil {
			return fmt.Errorf("failed to remove reaction role: %w", err)
		}

		return nil
	})
	if err != nil {
		h.log.Errorw("failed to handle reaction role", zap.Error(err))
	}
}

func (h *Handler) findEmoji(ctx context.Context, emoji discordgo.Emoji, messageID string) (*model.ReactionRole, error) {
	emojiName := emoji.Name
	if emoji.ID != "" {
		emojiName = fmt.Sprintf("%s:%s", emoji.Name, emoji.ID)
	}

	reaction, err := h.repo.FindByMessageAndEmoji(ctx, messageID, emojiName)
	if err != nil {
		return nil, fmt.Errorf("failed to find reaction role: %w", err)
	}
	return reaction, nil
}
