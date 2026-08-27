package main

import (
	"context"
	"os"

	"github.com/SkinonikS/discord-bot-go/internal/app/cli"
	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
	"github.com/pterm/pterm"
)

var (
	tag       = "dev"
	buildTime = "unknown"
	commit    = "none"
)

func main() {
	app, cmd := cli.NewApplication(foundation.BuildInfo{
		Tag:       tag,
		BuildTime: buildTime,
		Commit:    commit,
	})

	if err := app.Err(); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
