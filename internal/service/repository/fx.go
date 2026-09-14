package repository

import (
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/auto_role"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/disabled_guild_commands"
	pgsqlautorole "github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/auto_role"
	pgsqlguildcommandsetting "github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/disabled_guild_commands"
	pgsqlreactionrole "github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/reaction_role"
	pgsqlchannel "github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/temp_voice_channel"
	pgsqlchannelstate "github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/temp_voice_channel_state"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/reaction_role"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/temp_voice_channel"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/temp_voice_channel_state"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "repository"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(
			fx.Annotate(pgsqlchannel.NewRepo, fx.As(new(tempvoicechannel.Repo))),
			fx.Annotate(pgsqlchannelstate.NewRepo, fx.As(new(tempvoicechannelstate.Repo))),
			fx.Annotate(pgsqlreactionrole.NewRepo, fx.As(new(reactionrole.Repo))),
			fx.Annotate(pgsqlautorole.NewRepo, fx.As(new(autorole.Repo))),
			fx.Annotate(pgsqlguildcommandsetting.NewRepo, fx.As(new(guildcommandsetting.Repo))),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
