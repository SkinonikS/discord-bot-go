package main

import (
	"github.com/SkinonikS/discord-bot-go/internal/app/bot"
	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
)

var (
	tag       = "dev"
	buildTime = "unknown"
	commit    = "none"
)

func main() {
	app := bot.NewApplication(foundation.BuildInfo{
		Tag:       tag,
		BuildTime: buildTime,
		Commit:    commit,
	})
	app.Run()
}
