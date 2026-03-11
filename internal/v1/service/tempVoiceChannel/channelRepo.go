package tempVoiceChannel

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Channel struct {
	ID            uuid.UUID
	GuildID       snowflake.ID
	RootChannelID snowflake.ID
	ParentID      snowflake.ID
}

func (u *Channel) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	return nil
}

func (*Channel) TableName() string {
	return "temp_voice_channels"
}

type ChannelRepo interface {
	Transaction(ctx context.Context, fn func(tx ChannelRepo) error) error
	FindByCriteria(ctx context.Context, criteria ChannelSearchCriteria) ([]Channel, error)
	DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int, error)
	Save(ctx context.Context, setupChannel *Channel) error
}

type channelRepoImpl struct {
	db *gorm.DB
}

type ChannelRepoParams struct {
	fx.In

	DB *gorm.DB
}

func NewChannelRepo(p ChannelRepoParams) ChannelRepo {
	return &channelRepoImpl{
		db: p.DB,
	}
}

func (r *channelRepoImpl) Transaction(ctx context.Context, fn func(tx ChannelRepo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&channelRepoImpl{
			db: tx,
		})
	})
}

type ChannelSearchCriteria struct {
	GuildID       snowflake.ID
	RootChannelID snowflake.ID
	ParentID      snowflake.ID
}

func (r *channelRepoImpl) FindByCriteria(ctx context.Context, criteria ChannelSearchCriteria) ([]Channel, error) {
	var exprs []clause.Expression
	if criteria.GuildID != 0 {
		exprs = append(exprs, clause.Eq{Column: "guild_id", Value: criteria.GuildID})
	}
	if criteria.RootChannelID != 0 {
		exprs = append(exprs, clause.Eq{Column: "root_channel_id", Value: criteria.RootChannelID})
	}
	if criteria.ParentID != 0 {
		exprs = append(exprs, clause.Eq{Column: "parent_id", Value: criteria.ParentID})
	}

	return gorm.G[Channel](r.db, clause.Where{Exprs: exprs}).Find(ctx)
}

func (r *channelRepoImpl) Save(ctx context.Context, setupChannel *Channel) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "root_channel_id"}, {Name: "guild_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"parent_id"}),
		}).
		Create(setupChannel).Error
}

func (r *channelRepoImpl) DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	return gorm.G[*Channel](r.db).
		Where("id IN (?)", ids).
		Delete(ctx)
}
