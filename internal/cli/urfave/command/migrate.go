package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/alperdrsnn/clime"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
)

type MigrateParams struct {
	fx.In

	Migrator *goose.Provider
}

func NewMigrateCommand(p MigrateParams) *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Database migration commands",
		Commands: []*cli.Command{
			{
				Name:  "rollback",
				Usage: "Roll back the last applied migration",
				Action: func(ctx context.Context, _ *cli.Command) error {
					result, err := p.Migrator.Down(ctx)
					if err != nil {
						return fmt.Errorf("rollback failed: %w", err)
					}

					if result == nil {
						return errors.New("no migrations to roll back")
					}

					clime.SuccessLine("Rolled back: " + result.Source.Path)
					return nil
				},
			},
		},
	}
}
