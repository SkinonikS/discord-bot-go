package reactionRole

import (
	"context"
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
		OnGuildMessageReactionAdd:    el.GuildMessageReactionAdd,
		OnGuildMessageReactionRemove: el.GuildMessageReactionRemove,
		OnRoleDelete:                 el.RoleDelete,
		OnGuildMessageDelete:         el.GuildMessageDelete,
	}
}

func (el *eventListener) GuildMessageReactionAdd(e *disgoevents.GuildMessageReactionAdd) {
	err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if e.Member.User.Bot {
			return nil
		}

		return el.service.AssignRole(ctx, AssignRole{
			UserID:    e.Member.User.ID,
			GuildID:   e.GuildID,
			MessageID: e.MessageID,
			EmojiName: e.Emoji.Reaction(),
		})
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
			guildMember, err := e.Client().Rest.GetMember(
				e.GuildID,
				e.UserID,
				disgorest.WithCtx(ctx),
			)
			if err != nil {
				return fmt.Errorf("failed to get member: %w", err)
			}
			member = *guildMember
		}
		if member.User.Bot {
			return nil
		}

		return el.service.UnassignRole(ctx, UnassignRole{
			GuildID:   e.GuildID,
			MessageID: e.MessageID,
			EmojiName: e.Emoji.Reaction(),
			UserID:    e.UserID,
		})
	}); err != nil {
		el.log.Errorw("failed to handle reaction role", zap.Error(err))
	}
}

func (el *eventListener) RoleDelete(e *disgoevents.RoleDelete) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return el.service.UnassignOwnReactionsByRole(ctx, e.GuildID, e.RoleID)
	}); err != nil {
		el.log.Errorw("failed to handle role delete", zap.Error(err))
	}
}

func (el *eventListener) GuildMessageDelete(e *disgoevents.GuildMessageDelete) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := el.service.DeleteReactionsByMessage(ctx, e.GuildID, e.MessageID); err != nil {
			return fmt.Errorf("failed to delete reaction roles for deleted message: %w", err)
		}

		return nil
	}); err != nil {
		el.log.Errorw("failed to handle message delete", zap.Error(err))
	}
}
