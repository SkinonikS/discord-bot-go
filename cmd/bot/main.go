package main

import (
	"github.com/SkinonikS/discord-bot-go/internal/bot"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
)

var (
	tag       = "dev"
	buildTime = "unknown"
	commit    = "none"
)

func main() {
	buildInfo := foundation.NewBuildInfo(tag, buildTime, commit)
	app := bot.NewApplication(buildInfo)
	app.Run()
}
