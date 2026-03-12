package tempVoiceChannel

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChannelState struct {
	ID        uuid.UUID
	GuildID   snowflake.ID
	ChannelID snowflake.ID
}

func (u *ChannelState) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	return nil
}

func (*ChannelState) TableName() string {
	return "temp_voice_channel_states"
}

type ChannelStateRepo interface {
	Transaction(ctx context.Context, fn func(tx ChannelStateRepo) error) error
	FindByCriteria(ctx context.Context, criteria ChannelStateSearchCriteria) ([]ChannelState, error)
	DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int, error)
	Save(ctx context.Context, state *ChannelState) error
}

type channelStateRepoImpl struct {
	db            *gorm.DB
	discordConfig *discord.Config
}

type ChannelStateRepoParams struct {
	fx.In

	DB            *gorm.DB
	DiscordConfig *discord.Config
}

func NewChannelStateRepo(p ChannelStateRepoParams) ChannelStateRepo {
	return &channelStateRepoImpl{
		db:            p.DB,
		discordConfig: p.DiscordConfig,
	}
}

func (r *channelStateRepoImpl) Transaction(ctx context.Context, fn func(tx ChannelStateRepo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&channelStateRepoImpl{
			db: tx,
		})
	})
}

type ChannelStateSearchCriteria struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
}

func (r *channelStateRepoImpl) FindByCriteria(ctx context.Context, criteria ChannelStateSearchCriteria) ([]ChannelState, error) {
	var exprs []clause.Expression
	if criteria.GuildID != 0 {
		exprs = append(exprs, clause.Eq{Column: "guild_id", Value: criteria.GuildID})
	}
	if criteria.ChannelID != 0 {
		exprs = append(exprs, clause.Eq{Column: "channel_id", Value: criteria.ChannelID})
	}

	return gorm.G[ChannelState](r.db, clause.Where{Exprs: exprs}).Find(ctx)
}

func (r *channelStateRepoImpl) Save(ctx context.Context, state *ChannelState) error {
	return r.db.WithContext(ctx).Create(state).Error
}

func (r *channelStateRepoImpl) DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	return gorm.G[*ChannelState](r.db).
		Where("id IN (?)", ids).
		Delete(ctx)
}
