package interactionCommand

import (
	"context"

	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
)

type pingCommandImpl struct{}

func NewPingCommand() Command {
	return &pingCommandImpl{}
}

func (c *pingCommandImpl) Execute(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: "Pong!",
	}, disgorest.WithCtx(ctx))
}

func (c *pingCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:        c.Name(),
		Description: "Pong!",
	}
}

func (c *pingCommandImpl) Name() string {
	return "ping"
}
