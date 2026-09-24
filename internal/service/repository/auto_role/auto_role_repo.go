package autorole

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
)

type FindParams struct {
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

type DeleteParams struct {
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

type Repo interface {
	Find(ctx context.Context, params FindParams) ([]AutoRole, error)
	Delete(ctx context.Context, params DeleteParams) (int64, error)
	Save(ctx context.Context, autoRole *AutoRole) error
}
