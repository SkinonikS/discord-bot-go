package bot

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/cache"
	"github.com/SkinonikS/discord-bot-go/internal/infra/config"
	"github.com/SkinonikS/discord-bot-go/internal/infra/cron"
	"github.com/SkinonikS/discord-bot-go/internal/infra/database"
	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/infra/http_server"
	"github.com/SkinonikS/discord-bot-go/internal/infra/lavalink"
	"github.com/SkinonikS/discord-bot-go/internal/infra/logger"
	"github.com/SkinonikS/discord-bot-go/internal/infra/readiness"
	"github.com/SkinonikS/discord-bot-go/internal/infra/translator"
	"github.com/SkinonikS/discord-bot-go/internal/service"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewApplication(buildInfo foundation.BuildInfo) *fx.App {
	return fx.New(
		// INFRA
		foundation.NewModule(buildInfo),
		config.NewModule(),
		logger.NewModule(),
		database.NewModule(),
		cron.NewModule(),
		translator.NewModule(),
		discord.NewModule(),
		lavalink.NewModule(),
		http_server.NewModule(),
		readiness.NewModule(),
		cache.NewModule(),
		// SERVICE
		service.NewModule(),
		// LOGGER
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))}
		}),
	)
}
