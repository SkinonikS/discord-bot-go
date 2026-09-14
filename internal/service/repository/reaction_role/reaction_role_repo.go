package reactionrole

import (
	"context"

	"uuid"

	"github.com/disgoorg/snowflake/v2"
)

type SearchCriteria struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	MessageID snowflake.ID
	EmojiName string
	RoleID    snowflake.ID
}

type Repo interface {
	Transaction(ctx context.Context, fn func(tx Repo) error) error
	FindByCriteria(ctx context.Context, criteria SearchCriteria) ([]ReactionRole, error)
	DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int64, error)
	Save(ctx context.Context, reactionRole *ReactionRole) error
}
