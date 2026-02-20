package command

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/alperdrsnn/clime"
	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
)

type RegisterInteractionCommandsParams struct {
	fx.In
	InteractionCommands *interactionCommand.Registry
	Session             *discordgo.Session
	DiscordConfig       *discord.Config
}

func NewRegisterInteractionCommands(p RegisterInteractionCommandsParams) *cli.Command {
	return &cli.Command{
		Name:        "discord:interaction-commands:register",
		Description: "Register interaction commands in the guild.",
		Action: func(ctx context.Context, command *cli.Command) error {
			guildID := command.String("guild")
			if guildID == "" {
				return fmt.Errorf("guild ID is required")
			}

			clime.InfoLine("Registering interaction commands...")
			commands := p.InteractionCommands.List()
			definitions := make([]*discordgo.ApplicationCommand, 0, len(commands))
			for _, cmd := range commands {
				definitions = append(definitions, cmd.Definition())
			}

			_, err := p.Session.ApplicationCommandBulkOverwrite(p.DiscordConfig.AppID, guildID, definitions)
			if err != nil {
				return err
			}

			clime.SuccessLine("Successfully registered interaction commands")
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "guild",
				Aliases: []string{"g"},
				Usage:   "Guild ID",
			},
		},
	}
}
