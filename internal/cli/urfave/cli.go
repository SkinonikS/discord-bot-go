package urfave

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
)

type Params struct {
	fx.In
	AppConfig *config.Config
	Commands  []*cli.Command `group:"cli_commands"`
}

func New(p Params) *cli.Command {
	return &cli.Command{
		Name:     p.AppConfig.Name,
		Commands: p.Commands,
	}
}
