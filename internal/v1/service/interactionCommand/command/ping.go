package command

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type Ping struct {
}

func NewPing() *Ping {
	return &Ping{}
}

func (c *Ping) Execute(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate) error {
	if e.Type != discordgo.InteractionApplicationCommand {
		return nil
	}

	err := s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Pong!",
		},
	}, discordgo.WithContext(ctx))
	if err != nil {
		return err
	}

	return nil
}

func (c *Ping) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Pong!",
	}
}

func (c *Ping) ForOwnerOnly() bool {
	return false
}

func (c *Ping) Name() string {
	return "ping"
}
