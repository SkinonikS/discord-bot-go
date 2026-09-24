package autorole

import (
	"context"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	disgobot "github.com/disgoorg/disgo/bot"
	disgoevents "github.com/disgoorg/disgo/events"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type discordEventListener struct {
	service Service
	log     *zap.SugaredLogger
}

type DiscordEventListenerParams struct {
	fx.In

	Service Service
	Log     *zap.Logger
}

func NewDiscordEventListener(p DiscordEventListenerParams) disgobot.EventListener {
	el := &discordEventListener{
		log:     p.Log.Sugar(),
		service: p.Service,
	}

	return &disgoevents.ListenerAdapter{
		OnGuildMemberJoin: el.GuildMemberJoin,
		OnRoleDelete:      el.RoleDelete,
	}
}

func (el *discordEventListener) GuildMemberJoin(e *disgoevents.GuildMemberJoin) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if e.Member.User.Bot {
			return nil
		}

		return el.service.GuildMemberJoin(ctx, GuildMemberJoinParams{
			GuildID: e.GuildID,
			UserID:  e.Member.User.ID,
		})
	}); err != nil {
		el.log.Errorw("failed to handle auto role assignment", zap.Error(err))
	}
}

func (el *discordEventListener) RoleDelete(e *disgoevents.RoleDelete) {
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return el.service.RoleDelete(ctx, RoleDeleteParams{
			GuildID: e.GuildID,
			RoleID:  e.RoleID,
		})
	}); err != nil {
		el.log.Errorw("failed to handle role delete", zap.Error(err))
	}
}
