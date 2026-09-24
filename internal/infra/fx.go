package infraservice

import (
	"github.com/SkinonikS/discord-bot-go/internal/infra/cache"
	"github.com/SkinonikS/discord-bot-go/internal/infra/cli"
	"github.com/SkinonikS/discord-bot-go/internal/infra/config"
	"github.com/SkinonikS/discord-bot-go/internal/infra/cron"
	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
	httpserver "github.com/SkinonikS/discord-bot-go/internal/infra/http_server"
	"github.com/SkinonikS/discord-bot-go/internal/infra/lavalink"
	"github.com/SkinonikS/discord-bot-go/internal/infra/logger"
	"github.com/SkinonikS/discord-bot-go/internal/infra/migrator"
	"github.com/SkinonikS/discord-bot-go/internal/infra/postgres"
	"github.com/SkinonikS/discord-bot-go/internal/infra/readiness"
	"github.com/SkinonikS/discord-bot-go/internal/infra/redis"
	"github.com/SkinonikS/discord-bot-go/internal/infra/translator"
	"go.uber.org/fx"
)

const (
	ModuleName = "infra"
)

func NewModule(moduleParams foundation.ModuleParams) fx.Option {
	return fx.Module(ModuleName,
		foundation.NewModule(moduleParams),
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
		migrator.NewModule(),
		cron.NewModule(),
		cli.NewModule(),
	)
}
