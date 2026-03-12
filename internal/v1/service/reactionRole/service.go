package reactionRole

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrReactionRoleNotFound = errors.New("reaction role not found")
	ErrNotReactionRole      = errors.New("not reaction role")
)

type Service interface {
	GuildMessageReactionRemoveEmoji(ctx context.Context, params GuildMessageReactionRemoveEmoji) error
	GuildMessageReactionRemoveAll(ctx context.Context, params GuildMessageReactionRemoveAll) error
	GuildMessageDelete(ctx context.Context, params GuildMessageDelete) error
	RoleDelete(ctx context.Context, params RoleDelete) error
	GuildMessageReactionAdd(ctx context.Context, params GuildMessageReactionAdd) error
	GuildMessageReactionRemove(ctx context.Context, ur GuildMessageReactionRemove) error
	CreateReactionRole(ctx context.Context, params CreateReactionRole) (*ReactionRole, error)
	DeleteReactionRole(ctx context.Context, params DeleteReactionRole) error
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

type GuildMessageReactionRemoveEmoji struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	MessageID snowflake.ID
	EmojiName string
}

func (s *serviceImpl) GuildMessageReactionRemoveEmoji(ctx context.Context, params GuildMessageReactionRemoveEmoji) error {
	reactionRoles, err := s.repo.FindByCriteria(ctx, SearchCriteria{
		GuildID:   params.GuildID,
		ChannelID: params.ChannelID,
		MessageID: params.MessageID,
		EmojiName: params.EmojiName,
	})
	if err != nil {
		return err
	}

	_, err = s.repo.DeleteManyByIDs(ctx, lo.Map(reactionRoles, func(r ReactionRole, _ int) uuid.UUID {
		return r.ID
	}))
	return err
}

type GuildMessageReactionRemoveAll struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	MessageID snowflake.ID
}

func (s *serviceImpl) GuildMessageReactionRemoveAll(ctx context.Context, params GuildMessageReactionRemoveAll) error {
	reactionRoles, err := s.repo.FindByCriteria(ctx, SearchCriteria{
		GuildID:   params.GuildID,
		ChannelID: params.ChannelID,
		MessageID: params.MessageID,
	})
	if err != nil {
		return err
	}

	_, err = s.repo.DeleteManyByIDs(ctx, lo.Map(reactionRoles, func(r ReactionRole, _ int) uuid.UUID {
		return r.ID
	}))
	return err
}

type GuildMessageDelete struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	MessageID snowflake.ID
}

func (s *serviceImpl) GuildMessageDelete(ctx context.Context, params GuildMessageDelete) error {
	reactionRoles, err := s.repo.FindByCriteria(ctx, SearchCriteria{
		GuildID:   params.GuildID,
		ChannelID: params.ChannelID,
		MessageID: params.MessageID,
	})
	if err != nil {
		return err
	}

	_, err = s.repo.DeleteManyByIDs(ctx, lo.Map(reactionRoles, func(r ReactionRole, _ int) uuid.UUID {
		return r.ID
	}))
	return err
}

type RoleDelete struct {
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

func (s *serviceImpl) RoleDelete(ctx context.Context, params RoleDelete) error {
	reactionRoles, err := s.repo.FindByCriteria(ctx, SearchCriteria{
		GuildID: params.GuildID,
		RoleID:  params.RoleID,
	})
	if err != nil {
		return err
	}

	_, err = s.repo.DeleteManyByIDs(ctx, lo.Map(reactionRoles, func(r ReactionRole, _ int) uuid.UUID {
		return r.ID
	}))
	return err
}

type GuildMessageReactionAdd struct {
	GuildID   snowflake.ID
	MessageID snowflake.ID
	UserID    snowflake.ID
	EmojiName string
}

func (s *serviceImpl) GuildMessageReactionAdd(ctx context.Context, params GuildMessageReactionAdd) error {
	reactionRoles, err := s.repo.FindByCriteria(ctx, SearchCriteria{
		GuildID:   params.GuildID,
		MessageID: params.MessageID,
		EmojiName: params.EmojiName,
	})
	if err != nil {
		return err
	}
	if len(reactionRoles) == 0 {
		return ErrNotReactionRole
	}

	reactionRole := reactionRoles[0]
	if err := s.discordApi.AddMemberRole(params.GuildID, params.UserID, reactionRole.RoleID, disgorest.WithCtx(ctx)); err != nil {
		return fmt.Errorf("failed to add reaction role: %w", err)
	}

	return nil
}

type GuildMessageReactionRemove struct {
	GuildID   snowflake.ID
	MessageID snowflake.ID
	UserID    snowflake.ID
	EmojiName string
}

func (s *serviceImpl) GuildMessageReactionRemove(ctx context.Context, params GuildMessageReactionRemove) error {
	reactionRoles, err := s.repo.FindByCriteria(ctx, SearchCriteria{
		GuildID:   params.GuildID,
		MessageID: params.MessageID,
		EmojiName: params.EmojiName,
	})
	if err != nil {
		return err
	}
	if len(reactionRoles) == 0 {
		return ErrNotReactionRole
	}

	reactionRole := reactionRoles[0]
	if err := s.discordApi.RemoveMemberRole(params.GuildID, params.UserID, reactionRole.RoleID, disgorest.WithCtx(ctx)); err != nil {
		return fmt.Errorf("failed to remove reaction role: %w", err)
	}

	return nil
}

type CreateReactionRole struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	MessageID snowflake.ID
	EmojiName string
	RoleID    snowflake.ID
}

func (s *serviceImpl) CreateReactionRole(ctx context.Context, params CreateReactionRole) (*ReactionRole, error) {
	formattedEmoji := s.formatEmoji(params.EmojiName)
	reactionRole := &ReactionRole{
		GuildID:   params.GuildID,
		ChannelID: params.ChannelID,
		MessageID: params.MessageID,
		EmojiName: formattedEmoji,
		RoleID:    params.RoleID,
	}

	if err := s.repo.Transaction(ctx, func(tx Repo) error {
		if err := tx.Save(ctx, reactionRole); err != nil {
			return fmt.Errorf("failed to save reaction role: %w", err)
		}

		if err := s.discordApi.AddReaction(params.ChannelID, params.MessageID, formattedEmoji, disgorest.WithCtx(ctx)); err != nil {
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

func (s *serviceImpl) DeleteReactionRole(ctx context.Context, params DeleteReactionRole) error {
	formattedEmoji := s.formatEmoji(params.EmojiName)
	reactionRoles, err := s.repo.FindByCriteria(ctx, SearchCriteria{
		GuildID:   params.GuildID,
		MessageID: params.MessageID,
		EmojiName: formattedEmoji,
	})
	if err != nil {
		return fmt.Errorf("failed to find reaction role: %w", err)
	}
	if len(reactionRoles) == 0 {
		return ErrReactionRoleNotFound
	}

	reactionRole := reactionRoles[0]
	return s.repo.Transaction(ctx, func(tx Repo) error {
		if _, err := tx.DeleteManyByIDs(ctx, []uuid.UUID{reactionRole.ID}); err != nil {
			return fmt.Errorf("failed to remove reaction role: %w", err)
		}

		if err := s.discordApi.RemoveOwnReaction(reactionRole.ChannelID, params.MessageID, formattedEmoji, disgorest.WithCtx(ctx)); err != nil {
			return fmt.Errorf("failed to remove reaction from message: %w", err)
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
