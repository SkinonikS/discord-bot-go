package interactioncommand

import (
	"context"

	guildcommandsetting "github.com/SkinonikS/discord-bot-go/internal/service/repository/disabled_guild_commands"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

// Settings lets each guild independently enable or disable CommandScopeGuild commands.
type Settings interface {
	IsDisabled(ctx context.Context, guildID snowflake.ID, commandName string) (bool, error)
	ListDisabled(ctx context.Context, guildID snowflake.ID) ([]string, error)
	Disable(ctx context.Context, guildID snowflake.ID, commandName string) error
	Enable(ctx context.Context, guildID snowflake.ID, commandName string) error
}

type settingsImpl struct {
	repo guildcommandsetting.Repo
}

type SettingsParams struct {
	fx.In

	Repo guildcommandsetting.Repo
}

func NewSettings(p SettingsParams) Settings {
	return &settingsImpl{
		repo: p.Repo,
	}
}

func (s *settingsImpl) IsDisabled(ctx context.Context, guildID snowflake.ID, commandName string) (bool, error) {
	return s.repo.IsDisabled(ctx, guildID, commandName)
}

func (s *settingsImpl) ListDisabled(ctx context.Context, guildID snowflake.ID) ([]string, error) {
	return s.repo.ListDisabledByGuildID(ctx, guildID)
}

func (s *settingsImpl) Disable(ctx context.Context, guildID snowflake.ID, commandName string) error {
	return s.repo.Disable(ctx, guildID, commandName)
}

func (s *settingsImpl) Enable(ctx context.Context, guildID snowflake.ID, commandName string) error {
	return s.repo.Enable(ctx, guildID, commandName)
}
