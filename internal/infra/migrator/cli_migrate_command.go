package migrator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
)

const defaultStore = "postgres"

type CLIMigrateCommandParams struct {
	fx.In

	Registry *Registry
}

func NewCLIMigrateCommand(p CLIMigrateCommandParams) *cli.Command {
	return &cli.Command{
		Name:     "migrate",
		Usage:    "Database migration commands",
		Category: "MIGRATE",
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Apply all pending migrations for a store",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "store",
						Usage: "store to migrate, one of: " + strings.Join(p.Registry.List(), ", "),
						Value: defaultStore,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					store := cmd.String("store")

					provider, ok := p.Registry.Find(store)
					if !ok {
						return errors.New("store not found")
					}

					for {
						result, err := provider.UpByOne(ctx)
						if err != nil {
							if errors.Is(err, goose.ErrNoNextVersion) {
								break
							}
							return fmt.Errorf("migration failed: %w", err)
						}

						pterm.Success.Println("Applied " + pterm.Gray(result.Source.Path) + " for " + pterm.Gray(store) + " in " + pterm.Gray(result.Duration.String()))
					}

					pterm.Success.Println("Migrations for " + pterm.Gray(store) + " applied")
					return nil
				},
			},
			{
				Name:  "rollback",
				Usage: "Roll back the last applied migration for a store",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "store",
						Usage: "store to roll back, one of: " + strings.Join(p.Registry.List(), ", "),
						Value: defaultStore,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					store := cmd.String("store")

					provider, ok := p.Registry.Find(store)
					if !ok {
						return errors.New("store not found")
					}

					result, err := provider.Down(ctx)
					if err != nil {
						if errors.Is(err, goose.ErrNoNextVersion) {
							pterm.Info.Println("Nothing to rollback")
							return nil
						}

						return fmt.Errorf("rollback failed: %w", err)
					}

					pterm.Success.Println("Rolled back " + pterm.Gray(result.Source.Path) + " for " + pterm.Gray(store) + " in " + pterm.Gray(result.Duration.String()))
					return nil
				},
			},
		},
	}
}
