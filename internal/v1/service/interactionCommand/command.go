package interactionCommand

import (
	"context"

	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
)

type Command interface {
	Definition() disgodiscord.SlashCommandCreate
	Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error
	Name() string
}
