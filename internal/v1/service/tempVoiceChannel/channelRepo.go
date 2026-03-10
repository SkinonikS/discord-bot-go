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
	Find(ctx context.Context, guildID, channelID snowflake.ID) (*Channel, error)
	Save(ctx context.Context, setupChannel *Channel) error
	Delete(ctx context.Context, guildID, channelID snowflake.ID) (int, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (int, error)
	Transaction(ctx context.Context, fn func(tx ChannelRepo) error) error
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
		return fn(&channelRepoImpl{db: tx})
	})
}

func (r *channelRepoImpl) Find(
	ctx context.Context,
	guildID, channelID snowflake.ID,
) (*Channel, error) {
	setupChannel, err := gorm.G[*Channel](r.db).
		Where("guild_id = ?", guildID).
		Where("root_channel_id = ?", channelID).
		First(ctx)
	if err != nil {
		return nil, err
	}

	return setupChannel, nil
}

func (r *channelRepoImpl) Save(ctx context.Context, setupChannel *Channel) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "root_channel_id"}, {Name: "guild_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"parent_id"}),
		}).
		Create(setupChannel).Error
}

func (r *channelRepoImpl) Delete(
	ctx context.Context,
	guildID, channelID snowflake.ID,
) (int, error) {
	return gorm.G[*Channel](r.db).
		Where("guild_id = ?", guildID).
		Where("root_channel_id = ?", channelID).
		Delete(ctx)
}

func (r *channelRepoImpl) DeleteByID(ctx context.Context, id uuid.UUID) (int, error) {
	return gorm.G[*Channel](r.db).
		Where("id = ?", id).
		Delete(ctx)
}
