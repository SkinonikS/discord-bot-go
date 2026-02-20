package command

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type Command interface {
	Definition() *discordgo.ApplicationCommand
	Execute(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate) error
	ForOwnerOnly() bool
	Name() string
}
