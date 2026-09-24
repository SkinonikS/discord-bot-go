package interactioncommand

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	"github.com/SkinonikS/discord-bot-go/internal/infra/translator"
	disgobot "github.com/disgoorg/disgo/bot"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type discordEventListener struct {
	t        translator.Translator
	log      *zap.SugaredLogger
	commands Registry
	service  Service
}

type DiscordEventListenerParams struct {
	fx.In

	T        translator.Translator
	Log      *zap.Logger
	Commands Registry
	Service  Service
}

func NewDiscordEventListener(p DiscordEventListenerParams) disgobot.EventListener {
	el := &discordEventListener{
		t:        p.T,
		log:      p.Log.Sugar(),
		commands: p.Commands,
		service:  p.Service,
	}

	return &disgoevents.ListenerAdapter{
		OnApplicationCommandInteraction: el.ApplicationCommandInteractionCreate,
		OnGuildJoin:                     el.GuildJoin,
	}
}

func (el *discordEventListener) GuildJoin(e *disgoevents.GuildJoin) {
	const syncTimeout = 6 * time.Second
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), syncTimeout)
		defer cancel()

		if err := el.service.SyncGuildCommands(ctx, e.GuildID); err != nil {
			return fmt.Errorf("failed to sync guild commands: %w", err)
		}

		return nil
	}); err != nil {
		el.log.Errorw("failed to handle guild join", zap.Error(err))
	}
}

func (el *discordEventListener) ApplicationCommandInteractionCreate(e *disgoevents.ApplicationCommandInteractionCreate) {
	const defaultDiscordTimeout = 6 * time.Second
	if err := discord.ListenWithError(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultDiscordTimeout)
		defer cancel()

		if !el.isApplicable(e) {
			return nil
		}

		cmd, ok := el.commands.Find(e.Data.CommandName())
		if !ok {
			el.log.Warnw("unknown command executed", zap.String("command", e.Data.CommandName()))
			return nil
		}

		if err := cmd.Execute(ctx, e); err != nil {
			return fmt.Errorf("failed to handle command: %w", err)
		}

		return nil
	}); err != nil {
		el.notifyUserAboutError(e, err)
		el.log.Errorw("failed to handle interaction", zap.Error(err))
	}
}

func (el *discordEventListener) isApplicable(e *disgoevents.ApplicationCommandInteractionCreate) bool {
	return slices.Contains([]disgodiscord.InteractionType{
		disgodiscord.InteractionTypeApplicationCommand,
		disgodiscord.InteractionTypeAutocomplete,
	}, e.Type())
}

func (el *discordEventListener) notifyUserAboutError(e *disgoevents.ApplicationCommandInteractionCreate, err error) {
	if err := e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Embeds: []disgodiscord.Embed{
			{
				Title: el.t.SimpleLocalize(e.Locale(), "Execution Failed"),
				Description: el.t.Localize(e.Locale(), &i18n.LocalizeConfig{
					MessageID: "Something went wrong while executing this command.\n```{{.Error}}```",
					TemplateData: map[string]any{
						"Error": err.Error(),
					},
				}),
				Color: 0xff0000,
			},
		},
	}); err != nil {
		el.log.Errorw("unable to notify user about error", zap.Error(err))
	}
}
