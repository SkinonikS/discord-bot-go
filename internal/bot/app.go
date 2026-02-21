package bot

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"github.com/SkinonikS/discord-bot-go/internal/v1/cron"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database"
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/v1/logger"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewApplication(buildInfo *foundation.BuildInfo) *fx.App {
	return fx.New(
		foundation.NewModule(),
		config.NewModule(),
		logger.NewModule(),
		database.NewModule(),
		cron.NewModule(),
		discord.NewModule(),
		service.NewModule(),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))}
		}),
		fx.Supply(buildInfo),
	)
}
