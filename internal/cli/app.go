package cli

import (
	"github.com/SkinonikS/discord-bot-go/internal/cli/urfave"
	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"github.com/SkinonikS/discord-bot-go/internal/v1/cron"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database"
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/v1/logger"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service"
	"github.com/SkinonikS/discord-bot-go/internal/v1/translator"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewApplication(buildInfo *foundation.BuildInfo) (*fx.App, *cli.Command) {
	var cmd *cli.Command
	app := fx.New(
		foundation.NewModule(),
		config.NewModule(),
		logger.NewModule(),
		database.NewModule(),
		cron.NewModule(),
		urfave.NewModule(),
		translator.NewModule(),
		discord.NewModule(),
		service.NewModule(),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))}
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return zap.NewNop()
		}),
		fx.Populate(&cmd),
		fx.Supply(buildInfo),
	)
	return app, cmd
}
