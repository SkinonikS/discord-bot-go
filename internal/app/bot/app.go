package bot

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/cache"
	"github.com/SkinonikS/discord-bot-go/internal/infra/config"
	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
	httpserver "github.com/SkinonikS/discord-bot-go/internal/infra/http_server"
	"github.com/SkinonikS/discord-bot-go/internal/infra/lavalink"
	"github.com/SkinonikS/discord-bot-go/internal/infra/logger"
	"github.com/SkinonikS/discord-bot-go/internal/infra/postgres"
	"github.com/SkinonikS/discord-bot-go/internal/infra/readiness"
	"github.com/SkinonikS/discord-bot-go/internal/infra/redis"
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
		translator.NewModule(),
		discord.NewModule(),
		lavalink.NewModule(),
		httpserver.NewModule(),
		readiness.NewModule(),
		cache.NewModule(),
		postgres.NewModule(),
		redis.NewModule(),
		// SERVICE
		sharedservice.NewModule(),
		// LOGGER
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))}
		}),
	)
}
