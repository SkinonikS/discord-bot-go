package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/cli"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/alperdrsnn/clime"
)

var (
	tag       = "dev"
	buildTime = "unknown"
	commit    = "none"
)

func main() {
	buildInfo := foundation.NewBuildInfo(tag, buildTime, commit)
	app, cmd := cli.NewApplication(buildInfo)

	if err := app.Err(); err != nil {
		clime.ErrorLine(fmt.Sprintf("%v", err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := cmd.Run(ctx, os.Args); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			clime.ErrorLine("command execution timed out after 30 seconds")
			return
		}

		clime.ErrorLine(fmt.Sprintf("%v", err))
	}
}
