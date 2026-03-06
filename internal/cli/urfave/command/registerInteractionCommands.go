package command

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/alperdrsnn/clime"
	disgobot "github.com/disgoorg/disgo/bot"
	disgodiscord "github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
)

type RegisterInteractionCommandsParams struct {
	fx.In
	InteractionCommands *interactionCommand.Registry
	Client              *disgobot.Client
}

func NewRegisterInteractionCommands(p RegisterInteractionCommandsParams) *cli.Command {
	return &cli.Command{
		Name:        "discord:interaction-commands:register",
		Description: "Register interaction commands in the guild.",
		Action: func(ctx context.Context, command *cli.Command) error {
			guildIDStr := command.String("guild")
			if guildIDStr == "" {
				return fmt.Errorf("guild ID is required")
			}

			guildID, err := snowflake.Parse(guildIDStr)
			if err != nil {
				return fmt.Errorf("invalid guild ID: %w", err)
			}

			clime.InfoLine("Registering interaction commands...")
			commands := p.InteractionCommands.List()
			definitions := make([]disgodiscord.ApplicationCommandCreate, 0, len(commands))
			for _, cmd := range commands {
				definitions = append(definitions, cmd.Definition())
			}

			_, err = p.Client.Rest.SetGuildCommands(p.Client.ApplicationID, guildID, definitions)
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
