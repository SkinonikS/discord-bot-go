package autorole

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
)

type Repo interface {
	FindByGuildID(ctx context.Context, guildID snowflake.ID) ([]AutoRole, error)
	Save(ctx context.Context, autoRole *AutoRole) error
	DeleteByGuildIDAndRoleID(ctx context.Context, guildID, roleID snowflake.ID) (int64, error)
}
