package command

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/alperdrsnn/clime"
	"github.com/bwmarrin/discordgo"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
)

type LocalInteractionCommandsParams struct {
	fx.In
	InteractionCommands *interactionCommand.Registry
	DiscordConfig       *discord.Config
	Session             *discordgo.Session
}

func NewLocalInteractionCommands(p LocalInteractionCommandsParams) *cli.Command {
	return &cli.Command{
		Name:        "discord:interaction-commands:list",
		Description: "List interaction commands registered in the guild.",
		Action: func(ctx context.Context, command *cli.Command) error {
			guildID := command.String("guild")
			if guildID == "" {
				return fmt.Errorf("guild ID is required")
			}

			commands, err := p.Session.ApplicationCommands(p.DiscordConfig.AppID, guildID)
			if err != nil {
				return err
			}

			table := clime.NewTable().
				AddColumn("Name").
				AddColumn("Description").
				AddColumn("Registered?")

			for _, cmd := range p.InteractionCommands.List() {
				def := cmd.Definition()
				_, ok := lo.Find(commands, func(cmd *discordgo.ApplicationCommand) bool {
					return cmd.Name == def.Name
				})

				okStr := clime.Success.Sprintf("+")
				if !ok {
					okStr = clime.Error.Sprintf("-")
				}

				table.AddRow(def.Name, def.Description, okStr)
			}

			table.Print()
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
