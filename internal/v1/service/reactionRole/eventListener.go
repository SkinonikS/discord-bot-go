package reactionRole

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
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
		log:     p.Log.Sugar(),
		service: p.Service,
	}

	return &disgoevents.ListenerAdapter{
		OnGuildMessageReactionAdd:         el.GuildMessageReactionAdd,
		OnGuildMessageReactionRemove:      el.GuildMessageReactionRemove,
		OnRoleDelete:                      el.RoleDelete,
		OnGuildMessageDelete:              el.GuildMessageDelete,
		OnGuildMessageReactionRemoveEmoji: el.GuildMessageReactionRemoveEmoji,
		OnGuildMessageReactionRemoveAll:   el.GuildMessageReactionRemoveAll,
	}
}

func (el *eventListener) GuildMessageReactionRemoveEmoji(e *disgoevents.GuildMessageReactionRemoveEmoji) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return el.service.GuildMessageReactionRemoveEmoji(ctx, GuildMessageReactionRemoveEmoji{
			GuildID:   e.GuildID,
			ChannelID: e.ChannelID,
			MessageID: e.MessageID,
			EmojiName: e.Emoji.Reaction(),
		})
	}); err != nil {
		el.log.Errorw("failed to handle reaction role", zap.Error(err))
	}
}

func (el *eventListener) GuildMessageReactionRemoveAll(e *disgoevents.GuildMessageReactionRemoveAll) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return el.service.GuildMessageReactionRemoveAll(ctx, GuildMessageReactionRemoveAll{
			GuildID:   e.GuildID,
			ChannelID: e.ChannelID,
			MessageID: e.MessageID,
		})
	}); err != nil {
		el.log.Errorw("failed to handle reaction role", zap.Error(err))
	}
}

func (el *eventListener) RoleDelete(e *disgoevents.RoleDelete) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return el.service.RoleDelete(ctx, RoleDelete{
			GuildID: e.GuildID,
			RoleID:  e.RoleID,
		})
	}); err != nil {
		el.log.Errorw("failed to handle role delete", zap.Error(err))
	}
}

func (el *eventListener) GuildMessageReactionAdd(e *disgoevents.GuildMessageReactionAdd) {
	err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if e.Member.User.Bot {
			return nil
		}

		if err := el.service.GuildMessageReactionAdd(ctx, GuildMessageReactionAdd{
			UserID:    e.Member.User.ID,
			GuildID:   e.GuildID,
			MessageID: e.MessageID,
			EmojiName: e.Emoji.Reaction(),
		}); err != nil {
			if errors.Is(err, ErrNotReactionRole) {
				return nil
			}

			return err
		}

		return nil
	})
	if err != nil {
		el.log.Errorw("failed to handle reaction role", zap.Error(err))
	}
}

func (el *eventListener) GuildMessageReactionRemove(e *disgoevents.GuildMessageReactionRemove) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

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

		if err := el.service.GuildMessageReactionRemove(ctx, GuildMessageReactionRemove{
			GuildID:   e.GuildID,
			MessageID: e.MessageID,
			EmojiName: e.Emoji.Reaction(),
			UserID:    e.UserID,
		}); err != nil {
			if errors.Is(err, ErrNotReactionRole) {
				return nil
			}

			return err
		}

		return nil
	}); err != nil {
		el.log.Errorw("failed to handle reaction role", zap.Error(err))
	}
}

func (el *eventListener) GuildMessageDelete(e *disgoevents.GuildMessageDelete) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := el.service.GuildMessageDelete(ctx, GuildMessageDelete{
			GuildID:   e.GuildID,
			ChannelID: e.ChannelID,
			MessageID: e.MessageID,
		}); err != nil {
			return fmt.Errorf("failed to delete reaction roles for deleted message: %w", err)
		}

		return nil
	}); err != nil {
		el.log.Errorw("failed to handle message delete", zap.Error(err))
	}
}
