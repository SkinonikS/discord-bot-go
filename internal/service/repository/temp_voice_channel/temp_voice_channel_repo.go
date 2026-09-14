package tempvoicechannel

import (
	"context"
	"uuid"

	"github.com/disgoorg/snowflake/v2"
)

type SearchCriteria struct {
	GuildID       snowflake.ID
	RootChannelID snowflake.ID
	ParentID      snowflake.ID
}

type Repo interface {
	Transaction(ctx context.Context, fn func(tx Repo) error) error
	FindByCriteria(ctx context.Context, criteria SearchCriteria) ([]TempVoiceChannel, error)
	DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int64, error)
	Save(ctx context.Context, setupChannel *TempVoiceChannel) error
}
