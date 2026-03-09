package reactionRole

import (
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	"github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type eventListener struct {
	repo *repo.ReactionRoleRepo
	log  *zap.SugaredLogger
}

type EventListenerParams struct {
	fx.In
	Repo *repo.ReactionRoleRepo
	Log  *zap.Logger
}

func NewEventListener(p EventListenerParams) *disgoevents.ListenerAdapter {
	el := &eventListener{
		log:  p.Log.Sugar(),
		repo: p.Repo,
	}

	return &disgoevents.ListenerAdapter{
		OnGuildMessageReactionAdd:    el.GuildMessageReactionAdd,
		OnGuildMessageReactionRemove: el.GuildMessageReactionRemove,
		OnRoleDelete:                 el.RoleDelete,
		OnGuildMessageDelete:         el.GuildMessageDelete,
	}
}

func (el *eventListener) GuildMessageReactionAdd(e *disgoevents.GuildMessageReactionAdd) {
	err := discord.ListenWithError(func() error {
		ctx, cancel := discord.DefaultEventListenContext()
		defer cancel()

		if e.Member.User.Bot {
			return nil
		}

		reaction, err := el.repo.FindByMessageAndEmoji(ctx, e.MessageID, e.Emoji.Reaction())
		if err != nil {
			return err
		}
		if reaction == nil {
			return nil
		}

		if err := e.Client().Rest.AddMemberRole(e.GuildID, e.UserID, reaction.RoleID, disgorest.WithCtx(ctx)); err != nil {
			return fmt.Errorf("failed to add reaction role: %w", err)
		}

		return nil
	})
	if err != nil {
		el.log.Errorw("failed to handle reaction role", zap.Error(err))
	}
}

func (el *eventListener) GuildMessageReactionRemove(e *disgoevents.GuildMessageReactionRemove) {
	err := discord.ListenWithError(func() error {
		ctx, cancel := discord.DefaultEventListenContext()
		defer cancel()

		if e.Client().ID() == e.UserID {
			if err := el.repo.DeleteByMessageAndEmoji(ctx, e.MessageID, e.Emoji.Reaction()); err != nil {
				return fmt.Errorf("failed to delete reaction role after bot reaction removed: %w", err)
			}
			return nil
		}

		member, ok := e.Member()
		if !ok {
			guildMember, err := e.Client().Rest.GetMember(e.GuildID, e.UserID, disgorest.WithCtx(ctx))
			if err != nil {
				return fmt.Errorf("failed to get member: %w", err)
			}
			member = *guildMember
		}
		if member.User.Bot {
			return nil
		}

		reaction, err := el.repo.FindByMessageAndEmoji(ctx, e.MessageID, e.Emoji.Reaction())
		if err != nil {
			return err
		}
		if reaction == nil {
			return nil
		}

		if err := e.Client().Rest.RemoveMemberRole(e.GuildID, e.UserID, reaction.RoleID, disgorest.WithCtx(ctx)); err != nil {
			return fmt.Errorf("failed to remove reaction role: %w", err)
		}

		return nil
	})
	if err != nil {
		el.log.Errorw("failed to handle reaction role", zap.Error(err))
	}
}

func (el *eventListener) RoleDelete(e *disgoevents.RoleDelete) {
	err := discord.ListenWithError(func() error {
		ctx, cancel := discord.DefaultEventListenContext()
		defer cancel()

		records, err := el.repo.FindAllByRoleID(ctx, e.RoleID)
		if err != nil {
			return fmt.Errorf("failed to find reaction roles for deleted role: %w", err)
		}

		for _, record := range records {
			if err := e.Client().Rest.RemoveOwnReaction(record.ChannelID, record.MessageID, record.EmojiName, disgorest.WithCtx(ctx)); err != nil {
				el.log.Warnw("failed to remove bot reaction for deleted role", "channelID", record.ChannelID, "messageID", record.MessageID, zap.Error(err))
			}
		}

		if err := el.repo.DeleteByRoleID(ctx, e.RoleID); err != nil {
			return fmt.Errorf("failed to delete reaction roles for deleted role: %w", err)
		}

		return nil
	})
	if err != nil {
		el.log.Errorw("failed to handle role delete", zap.Error(err))
	}
}

func (el *eventListener) GuildMessageDelete(e *disgoevents.GuildMessageDelete) {
	err := discord.ListenWithError(func() error {
		ctx, cancel := discord.DefaultEventListenContext()
		defer cancel()

		if err := el.repo.DeleteByMessageID(ctx, e.MessageID); err != nil {
			return fmt.Errorf("failed to delete reaction roles for deleted message: %w", err)
		}
		return nil
	})
	if err != nil {
		el.log.Errorw("failed to handle message delete", zap.Error(err))
	}
}
