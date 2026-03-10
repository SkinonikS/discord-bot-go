package reactionRole

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrReactionRoleNotFound = errors.New("reaction role not found")

type Service interface {
	CreateReactionRole(ctx context.Context, cr CreateReactionRole) (*ReactionRole, error)
	DeleteReactionRole(ctx context.Context, dr DeleteReactionRole) error
	AssignRole(ctx context.Context, ar AssignRole) error
	UnassignRole(ctx context.Context, ur UnassignRole) error
	UnassignOwnReactionsByRole(ctx context.Context, guildID, roleID snowflake.ID) error
	DeleteReactionsByMessage(ctx context.Context, guildID, messageID snowflake.ID) error
	DeleteReactionsByMessageAndEmoji(ctx context.Context, dr DeleteReactionsByMessageAndEmoji) error
	DeleteStaleReactions(ctx context.Context, shardID uint32) error
}

type serviceImpl struct {
	emojiRegex *regexp.Regexp
	repo       Repo
	discordApi disgorest.Rest
	log        *zap.SugaredLogger
}

type ServiceParams struct {
	fx.In

	Log        *zap.Logger
	DiscordApi disgorest.Rest
	Repo       Repo
}

func NewService(p ServiceParams) Service {
	return &serviceImpl{
		emojiRegex: regexp.MustCompile(`^<(a?):(\w+):(\d+)>$`),
		discordApi: p.DiscordApi,
		repo:       p.Repo,
		log:        p.Log.Sugar(),
	}
}

type CreateReactionRole struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	MessageID snowflake.ID
	EmojiName string
	RoleID    snowflake.ID
}

func (s *serviceImpl) CreateReactionRole(
	ctx context.Context,
	cr CreateReactionRole,
) (*ReactionRole, error) {
	formattedEmoji := s.formatEmoji(cr.EmojiName)
	reactionRole := &ReactionRole{
		GuildID:   cr.GuildID,
		ChannelID: cr.ChannelID,
		MessageID: cr.MessageID,
		EmojiName: formattedEmoji,
		RoleID:    cr.RoleID,
	}

	if err := s.repo.Transaction(ctx, func(tx Repo) error {
		if err := tx.Save(ctx, reactionRole); err != nil {
			return fmt.Errorf("failed to save reaction role: %w", err)
		}

		if err := s.discordApi.AddReaction(
			cr.ChannelID,
			cr.MessageID,
			formattedEmoji,
			disgorest.WithCtx(ctx),
		); err != nil {
			return fmt.Errorf("failed to add reaction to message: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return reactionRole, nil
}

type DeleteReactionRole struct {
	GuildID   snowflake.ID
	MessageID snowflake.ID
	EmojiName string
}

func (s *serviceImpl) DeleteReactionRole(ctx context.Context, dr DeleteReactionRole) error {
	formattedEmoji := s.formatEmoji(dr.EmojiName)
	reaction, err := s.repo.FindByMessageAndEmoji(ctx, dr.GuildID, dr.MessageID, formattedEmoji)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReactionRoleNotFound
		}
		return fmt.Errorf("failed to find reaction role: %w", err)
	}

	return s.repo.Transaction(ctx, func(tx Repo) error {
		if _, err := tx.DeleteByID(ctx, reaction.ID); err != nil {
			return fmt.Errorf("failed to remove reaction role: %w", err)
		}

		if err := s.discordApi.RemoveOwnReaction(
			reaction.ChannelID,
			dr.MessageID,
			formattedEmoji,
			disgorest.WithCtx(ctx),
		); err != nil {
			return fmt.Errorf("failed to remove reaction from message: %w", err)
		}

		return nil
	})
}

type AssignRole struct {
	GuildID   snowflake.ID
	MessageID snowflake.ID
	UserID    snowflake.ID
	EmojiName string
}

func (s *serviceImpl) AssignRole(ctx context.Context, ar AssignRole) error {
	reaction, err := s.repo.FindByMessageAndEmoji(ctx, ar.GuildID, ar.MessageID, ar.EmojiName)
	if err != nil {
		return err
	}
	if reaction == nil {
		return nil
	}

	if err := s.discordApi.AddMemberRole(
		ar.GuildID,
		ar.UserID,
		reaction.RoleID,
		disgorest.WithCtx(ctx),
	); err != nil {
		return fmt.Errorf("failed to add reaction role: %w", err)
	}

	return nil
}

type UnassignRole struct {
	GuildID   snowflake.ID
	MessageID snowflake.ID
	UserID    snowflake.ID
	EmojiName string
}

func (s *serviceImpl) UnassignRole(ctx context.Context, ur UnassignRole) error {
	reaction, err := s.repo.FindByMessageAndEmoji(ctx, ur.GuildID, ur.MessageID, ur.EmojiName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return err
	}

	if err := s.discordApi.RemoveMemberRole(
		ur.GuildID,
		ur.UserID,
		reaction.RoleID,
		disgorest.WithCtx(ctx),
	); err != nil {
		return fmt.Errorf("failed to remove reaction role: %w", err)
	}

	return nil
}

func (s *serviceImpl) UnassignOwnReactionsByRole(
	ctx context.Context,
	guildID, roleID snowflake.ID,
) error {
	records, err := s.repo.FindAllByRoleID(ctx, guildID, roleID)
	if err != nil {
		return fmt.Errorf("failed to find reaction roles for deleted role: %w", err)
	}

	return s.repo.Transaction(ctx, func(tx Repo) error {
		for _, record := range records {
			if err := s.discordApi.RemoveOwnReaction(
				record.ChannelID,
				record.MessageID,
				record.EmojiName,
				disgorest.WithCtx(ctx),
			); err != nil {
				return fmt.Errorf("failed to remove bot reaction for deleted role: %w", err)
			}

			time.Sleep(100 * time.Millisecond)
		}

		if _, err := tx.DeleteByRoleID(ctx, guildID, roleID); err != nil {
			return fmt.Errorf("failed to delete reaction roles for deleted role: %w", err)
		}

		return nil
	})
}

func (s *serviceImpl) DeleteReactionsByMessage(
	ctx context.Context,
	guildID, messageID snowflake.ID,
) error {
	if _, err := s.repo.DeleteByMessageID(ctx, guildID, messageID); err != nil {
		return fmt.Errorf("failed to delete reaction roles for deleted message: %w", err)
	}

	return nil
}

type DeleteReactionsByMessageAndEmoji struct {
	GuildID   snowflake.ID
	MessageID snowflake.ID
	Emoji     string
}

func (s *serviceImpl) DeleteReactionsByMessageAndEmoji(
	ctx context.Context,
	dr DeleteReactionsByMessageAndEmoji,
) error {
	if _, err := s.repo.DeleteByMessageAndEmoji(
		ctx,
		dr.GuildID,
		dr.MessageID,
		dr.Emoji,
	); err != nil {
		return fmt.Errorf("failed to delete reaction roles for deleted message and emoji: %w", err)
	}

	return nil
}

func (s *serviceImpl) DeleteStaleReactions(ctx context.Context, shardID uint32) error {
	const batchSize = 100
	return s.repo.FindInBatchesByShard(ctx, shardID, batchSize, func(records []ReactionRole, _ int) error {
		type msgKey struct {
			roleID    snowflake.ID
			channelID snowflake.ID
			messageID snowflake.ID
			guildID   snowflake.ID
		}

		groups := make(map[msgKey][]ReactionRole, len(records))
		for _, r := range records {
			key := msgKey{
				roleID:    r.RoleID,
				channelID: r.ChannelID,
				messageID: r.MessageID,
				guildID:   r.GuildID,
			}
			groups[key] = append(groups[key], r)
		}

		for key, roles := range groups {
			msg, err := s.discordApi.GetMessage(key.channelID, key.messageID, disgorest.WithCtx(ctx))
			if err != nil {
				if restErr, ok := errors.AsType[*disgorest.Error](err); ok && restErr.Response != nil && restErr.Response.StatusCode == http.StatusNotFound {
					if _, err := s.repo.DeleteByMessageID(ctx, key.guildID, key.messageID); err != nil {
						s.log.Warnw("failed to delete records for missing message", zap.Error(err))
					}
				}
				continue
			}

			botReacted := make(map[string]bool, len(msg.Reactions))
			for i := range msg.Reactions {
				r := &msg.Reactions[i]
				if r.Me {
					botReacted[r.Emoji.Reaction()] = true
				}
			}

			for _, role := range roles {
				if !botReacted[role.EmojiName] {
					if _, err := s.repo.DeleteByMessageAndEmoji(ctx, role.GuildID, role.MessageID, role.EmojiName); err != nil {
						s.log.Warnw("failed to delete orphaned reaction role", zap.Error(err))
					}
				}
			}

			time.Sleep(1 * time.Second)
		}

		return nil
	})
}

func (s *serviceImpl) formatEmoji(emoji string) string {
	if matches := s.emojiRegex.FindStringSubmatch(emoji); len(matches) == 4 {
		return fmt.Sprintf("%s:%s", matches[2], matches[3])
	}

	return strings.TrimSpace(emoji)
}
