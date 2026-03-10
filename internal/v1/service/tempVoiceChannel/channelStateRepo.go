package tempVoiceChannel

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/scope"
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"gorm.io/gorm"
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
	Find(ctx context.Context, guildID, channelID snowflake.ID) (*ChannelState, error)
	Save(ctx context.Context, state *ChannelState) error
	Delete(ctx context.Context, guildID, channelID snowflake.ID) (int, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (int, error)
	FindInBatchesByShard(
		ctx context.Context,
		shardID uint32,
		batchSize int,
		cb func([]ChannelState, int) error,
	) error
	Transaction(ctx context.Context, fn func(tx ChannelStateRepo) error) error
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

func (r *channelStateRepoImpl) Transaction(
	ctx context.Context,
	fn func(tx ChannelStateRepo) error,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&channelStateRepoImpl{db: tx})
	})
}

func (r *channelStateRepoImpl) Find(
	ctx context.Context,
	guildID, channelID snowflake.ID,
) (*ChannelState, error) {
	state, err := gorm.G[*ChannelState](r.db).
		Where("guild_id = ?", guildID).
		Where("channel_id = ?", channelID).
		First(ctx)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (r *channelStateRepoImpl) Save(ctx context.Context, state *ChannelState) error {
	return r.db.WithContext(ctx).Create(state).Error
}

func (r *channelStateRepoImpl) Delete(
	ctx context.Context,
	guildID, channelID snowflake.ID,
) (int, error) {
	return gorm.G[*ChannelState](r.db).
		Where("guild_id = ?", guildID).
		Where("channel_id = ?", channelID).
		Delete(ctx)
}

func (r *channelStateRepoImpl) DeleteByID(ctx context.Context, id uuid.UUID) (int, error) {
	return gorm.G[*ChannelState](r.db).
		Where("id = ?", id).
		Delete(ctx)
}

func (r *channelStateRepoImpl) FindInBatchesByShard(
	ctx context.Context,
	shardID uint32,
	batchSize int,
	cb func([]ChannelState, int) error,
) error {
	return gorm.G[ChannelState](r.db).
		Scopes(scope.ByShard(shardID, r.discordConfig.ShardCount)).
		FindInBatches(ctx, batchSize, cb)
}
