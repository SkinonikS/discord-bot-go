package reactionRole

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/scope"
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReactionRole struct {
	ID        uuid.UUID
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	MessageID snowflake.ID
	EmojiName string
	RoleID    snowflake.ID
}

func (u *ReactionRole) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (*ReactionRole) TableName() string {
	return "reaction_roles"
}

type Repo interface {
	Transaction(ctx context.Context, fn func(tx Repo) error) error
	FindByMessageAndEmoji(ctx context.Context, guildID, messageID snowflake.ID, emojiName string) (*ReactionRole, error)
	FindAllByRoleID(ctx context.Context, guildID, roleID snowflake.ID) ([]ReactionRole, error)
	FindInBatchesByShard(ctx context.Context, shardID uint32, batchSize int, cb func([]ReactionRole, int) error) error
	DeleteByID(ctx context.Context, id uuid.UUID) (int, error)
	DeleteByEmoji(ctx context.Context, guildID snowflake.ID, emojiName string) (int, error)
	DeleteByRoleID(ctx context.Context, guildID, roleID snowflake.ID) (int, error)
	DeleteByMessageID(ctx context.Context, guildID, messageID snowflake.ID) (int, error)
	DeleteByMessageAndEmoji(ctx context.Context, guildID, messageID snowflake.ID, emojiName string) (int, error)
	Save(ctx context.Context, reactionRole *ReactionRole) error
}

type repoImpl struct {
	db            *gorm.DB
	discordConfig *discord.Config
}

type RepoParams struct {
	fx.In

	DB            *gorm.DB
	DiscordConfig *discord.Config
}

func NewRepo(p RepoParams) Repo {
	return &repoImpl{
		discordConfig: p.DiscordConfig,
		db:            p.DB,
	}
}

func (r *repoImpl) Transaction(ctx context.Context, fn func(tx Repo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&repoImpl{
			db:            tx,
			discordConfig: r.discordConfig,
		})
	})
}

func (r *repoImpl) Save(ctx context.Context, reactionRole *ReactionRole) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}, {Name: "emoji_name"}, {Name: "guild_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"role_id"}),
		}).
		Create(reactionRole).Error
}

func (r *repoImpl) FindByMessageAndEmoji(
	ctx context.Context,
	guildID, messageID snowflake.ID,
	emojiName string,
) (*ReactionRole, error) {
	role, err := gorm.G[*ReactionRole](r.db).
		Where("guild_id = ?", guildID).
		Where("message_id = ?", messageID).
		Where("emoji_name = ?", emojiName).
		First(ctx)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r *repoImpl) FindAllByRoleID(
	ctx context.Context,
	guildID, roleID snowflake.ID,
) ([]ReactionRole, error) {
	return gorm.G[ReactionRole](r.db).
		Where("guild_id = ?", guildID).
		Where("role_id = ?", roleID).
		Find(ctx)
}

func (r *repoImpl) FindInBatchesByShard(
	ctx context.Context,
	shardID uint32,
	batchSize int,
	cb func([]ReactionRole, int) error,
) error {
	return gorm.G[ReactionRole](r.db).
		Scopes(scope.ByShard(shardID, r.discordConfig.ShardCount)).
		FindInBatches(ctx, batchSize, cb)
}

func (r *repoImpl) DeleteByEmoji(
	ctx context.Context,
	guildID snowflake.ID,
	emojiName string,
) (int, error) {
	return gorm.G[ReactionRole](r.db).
		Where("guild_id = ?", guildID).
		Where("emoji_name = ?", emojiName).
		Delete(ctx)
}

func (r *repoImpl) DeleteByID(ctx context.Context, id uuid.UUID) (int, error) {
	return gorm.G[*ReactionRole](r.db).
		Where("id = ?", id).
		Delete(ctx)
}

func (r *repoImpl) DeleteByMessageAndEmoji(
	ctx context.Context,
	guildID, messageID snowflake.ID,
	emojiName string,
) (int, error) {
	return gorm.G[*ReactionRole](r.db).
		Where("guild_id = ?", guildID).
		Where("message_id = ?", messageID).
		Where("emoji_name = ?", emojiName).
		Delete(ctx)
}

func (r *repoImpl) DeleteByMessageID(
	ctx context.Context,
	guildID, messageID snowflake.ID,
) (int, error) {
	return gorm.G[*ReactionRole](r.db).
		Where("guild_id = ?", guildID).
		Where("message_id = ?", messageID).
		Delete(ctx)
}

func (r *repoImpl) DeleteByRoleID(ctx context.Context, guildID, roleID snowflake.ID) (int, error) {
	return gorm.G[*ReactionRole](r.db).
		Where("guild_id = ?", guildID).
		Where("role_id = ?", roleID).
		Delete(ctx)
}
