package cli

import (
	"github.com/SkinonikS/discord-bot-go/internal/app/cli/internal"
	"github.com/SkinonikS/discord-bot-go/internal/infra/cache"
	"github.com/SkinonikS/discord-bot-go/internal/infra/config"
	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/infra/lavalink"
	"github.com/SkinonikS/discord-bot-go/internal/infra/logger"
	"github.com/SkinonikS/discord-bot-go/internal/infra/migrator"
	"github.com/SkinonikS/discord-bot-go/internal/infra/postgres"
	"github.com/SkinonikS/discord-bot-go/internal/infra/translator"
	"github.com/SkinonikS/discord-bot-go/internal/service"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewApplication(buildInfo foundation.BuildInfo) (*fx.App, *cli.Command) {
	var cmd *cli.Command
	app := fx.New(
		// INFRA
		foundation.NewModule(buildInfo),
		config.NewModule(),
		logger.NewModule(),
		translator.NewModule(),
		lavalink.NewModule(),
		discord.NewModule(),
		cache.NewModule(),
		postgres.NewModule(),
		migrator.NewModule(),
		// SERVICE
		sharedservice.NewModule(),
		appservice.NewModule(),
		// LOGGER
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))}
		}),
		fx.Decorate(func(*zap.Logger) *zap.Logger {
			return zap.NewNop()
		}),
		// ROOT CMD
		fx.Populate(&cmd),
	)
	return app, cmd
}
