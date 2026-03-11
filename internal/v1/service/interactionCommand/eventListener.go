package interactionCommand

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type eventListener struct {
	log      *zap.SugaredLogger
	commands Registry
}

type EventListenerParams struct {
	fx.In

	Log      *zap.Logger
	Commands Registry
}

func NewEventListener(p EventListenerParams) *disgoevents.ListenerAdapter {
	el := &eventListener{
		commands: p.Commands,
		log:      p.Log.Sugar(),
	}

	return &disgoevents.ListenerAdapter{
		OnApplicationCommandInteraction: el.ApplicationCommandInteractionCreate,
	}
}

func (el *eventListener) ApplicationCommandInteractionCreate(e *disgoevents.ApplicationCommandInteractionCreate) {
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

func (el *eventListener) isApplicable(e *disgoevents.ApplicationCommandInteractionCreate) bool {
	return slices.Contains([]disgodiscord.InteractionType{
		disgodiscord.InteractionTypeApplicationCommand,
		disgodiscord.InteractionTypeAutocomplete,
	}, e.Type())
}

func (el *eventListener) notifyUserAboutError(e *disgoevents.ApplicationCommandInteractionCreate, err error) {
	if err := e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Embeds: []disgodiscord.Embed{
			{
				Title:       "Execution Failed",
				Description: "Something went wrong while executing this command.\n```" + err.Error() + "```",
				Color:       0xff0000,
			},
		},
	}); err != nil {
		el.log.Errorw("unable to notify user about error", zap.Error(err))
	}
}
