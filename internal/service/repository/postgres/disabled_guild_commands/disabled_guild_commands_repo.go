package disabled_guild_commands

import (
	"context"

	guildcommandsetting "github.com/SkinonikS/discord-bot-go/internal/service/repository/disabled_guild_commands"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/internal/gen"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var _ guildcommandsetting.Repo = (*Repo)(nil)

type Repo struct {
	queries *gen.Queries
}

type Params struct {
	fx.In

	Pool *pgxpool.Pool
}

func NewRepo(p Params) *Repo {
	return &Repo{
		queries: gen.New(p.Pool),
	}
}

func (r *Repo) IsDisabled(ctx context.Context, guildID snowflake.ID, commandName string) (bool, error) {
	return r.queries.IsGuildCommandDisabled(ctx, gen.IsGuildCommandDisabledParams{
		GuildID:     guildID,
		CommandName: commandName,
	})
}

func (r *Repo) ListDisabledByGuildID(ctx context.Context, guildID snowflake.ID) ([]string, error) {
	return r.queries.ListDisabledGuildCommandsByGuildID(ctx, guildID)
}

func (r *Repo) Disable(ctx context.Context, guildID snowflake.ID, commandName string) error {
	return r.queries.DisableGuildCommand(ctx, gen.DisableGuildCommandParams{
		GuildID:     guildID,
		CommandName: commandName,
	})
}

func (r *Repo) Enable(ctx context.Context, guildID snowflake.ID, commandName string) error {
	return r.queries.EnableGuildCommand(ctx, gen.EnableGuildCommandParams{
		GuildID:     guildID,
		CommandName: commandName,
	})
}
