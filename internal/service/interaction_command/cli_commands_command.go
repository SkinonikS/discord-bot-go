package interactioncommand

import (
	"context"
	"fmt"
	"sort"

	"github.com/disgoorg/snowflake/v2"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
)

type CLICommandsCommand struct {
	fx.In

	Registry Registry
	Service  Service
}

func NewCLICommandsCommand(p CLICommandsCommand) *cli.Command {
	return &cli.Command{
		Name:     "commands",
		Usage:    "Discord slash command registration",
		Category: "DISCORD COMMANDS",
		Commands: []*cli.Command{
			{
				Name:  "sync",
				Usage: "Register global commands and guild commands for every guild the bot is currently in",
				Action: func(ctx context.Context, _ *cli.Command) error {
					if err := p.Service.SyncGlobalCommands(ctx); err != nil {
						return fmt.Errorf("sync global commands failed: %w", err)
					}

					if err := p.Service.SyncAllGuildCommands(ctx); err != nil {
						return fmt.Errorf("sync guild commands failed: %w", err)
					}

					pterm.Success.Println("Slash commands synced")
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "List guild-scoped commands and whether they're enabled for a given guild",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "guild",
						Usage:    "ID of the guild to list commands for",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					guildID, err := snowflake.Parse(cmd.String("guild"))
					if err != nil {
						return fmt.Errorf("invalid guild ID: %w", err)
					}

					disabled, err := p.Service.ListDisabledGuildCommands(ctx, guildID)
					if err != nil {
						return fmt.Errorf("failed to list disabled commands: %w", err)
					}
					disabledSet := lo.SliceToMap(disabled, func(name string) (string, struct{}) {
						return name, struct{}{}
					})

					commands := p.Registry.ListByScope(CommandScopeGuild)
					sort.Slice(commands, func(i, j int) bool {
						return commands[i].Name() < commands[j].Name()
					})

					rows := [][]string{{"Command", "Status"}}
					for _, c := range commands {
						status := "enabled"
						if _, ok := disabledSet[c.Name()]; ok {
							status = "disabled"
						}

						rows = append(rows, []string{"/" + c.Name(), status})
					}

					return pterm.DefaultTable.WithHasHeader().WithData(rows).Render()
				},
			},
		},
	}
}
