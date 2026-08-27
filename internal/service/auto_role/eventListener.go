package auto_role

import (
	"context"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	disgobot "github.com/disgoorg/disgo/bot"
	disgoevents "github.com/disgoorg/disgo/events"
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

func NewEventListener(p EventListenerParams) disgobot.EventListener {
	el := &eventListener{
		log:     p.Log.Sugar(),
		service: p.Service,
	}

	return &disgoevents.ListenerAdapter{
		OnGuildMemberJoin: el.GuildMemberJoin,
		OnRoleDelete:      el.RoleDelete,
	}
}

func (el *eventListener) GuildMemberJoin(e *disgoevents.GuildMemberJoin) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if e.Member.User.Bot {
			return nil
		}

		return el.service.GuildMemberJoin(ctx, GuildMemberJoin{
			GuildID: e.GuildID,
			UserID:  e.Member.User.ID,
		})
	}); err != nil {
		el.log.Errorw("failed to handle auto role assignment", zap.Error(err))
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
