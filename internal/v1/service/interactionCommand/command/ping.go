package command

import (
	"context"

	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
)

type Ping struct {
}

func NewPing() *Ping {
	return &Ping{}
}

func (c *Ping) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: "Pong!",
	}, disgorest.WithCtx(ctx))
}

func (c *Ping) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:        c.Name(),
		Description: "Pong!",
	}
}

func (c *Ping) Name() string {
	return "ping"
}
