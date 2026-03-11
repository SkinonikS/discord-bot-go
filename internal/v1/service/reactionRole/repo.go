package reactionRole

import (
	"context"

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
	FindByCriteria(ctx context.Context, criteria SearchCriteria) ([]ReactionRole, error)
	DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int, error)
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

type SearchCriteria struct {
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	MessageID snowflake.ID
	EmojiName string
	RoleID    snowflake.ID
}

func (r *repoImpl) FindByCriteria(ctx context.Context, criteria SearchCriteria) ([]ReactionRole, error) {
	var exprs []clause.Expression
	if criteria.GuildID != 0 {
		exprs = append(exprs, clause.Eq{Column: "guild_id", Value: criteria.GuildID})
	}
	if criteria.ChannelID != 0 {
		exprs = append(exprs, clause.Eq{Column: "channel_id", Value: criteria.ChannelID})
	}
	if criteria.MessageID != 0 {
		exprs = append(exprs, clause.Eq{Column: "message_id", Value: criteria.MessageID})
	}
	if criteria.EmojiName != "" {
		exprs = append(exprs, clause.Eq{Column: "emoji_name", Value: criteria.EmojiName})
	}
	if criteria.RoleID != 0 {
		exprs = append(exprs, clause.Eq{Column: "role_id", Value: criteria.RoleID})
	}

	return gorm.G[ReactionRole](r.db, clause.Where{Exprs: exprs}).Find(ctx)
}

func (r *repoImpl) DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	return gorm.G[ReactionRole](r.db).
		Where("id IN (?)", ids).
		Delete(ctx)
}
