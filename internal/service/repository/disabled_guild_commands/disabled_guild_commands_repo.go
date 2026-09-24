package guildcommandsetting

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
)

type Repo interface {
	IsDisabled(ctx context.Context, guildId snowflake.ID, commandName string) (bool, error)
	ListDisabled(ctx context.Context, guildID snowflake.ID) ([]string, error)
	Disable(ctx context.Context, guildId snowflake.ID, commandName string) error
	Enable(ctx context.Context, guildId snowflake.ID, commandName string) error
}
