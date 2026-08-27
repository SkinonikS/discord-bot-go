package auto_role

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AutoRole struct {
	ID      uuid.UUID `gorm:"primaryKey"`
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

func (u *AutoRole) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (*AutoRole) TableName() string {
	return "auto_roles"
}

type Repo interface {
	FindByGuildID(ctx context.Context, guildID snowflake.ID) ([]AutoRole, error)
	Save(ctx context.Context, autoRole *AutoRole) error
	DeleteByGuildIDAndRoleID(ctx context.Context, guildID, roleID snowflake.ID) (int, error)
}

type repoImpl struct {
	db *gorm.DB
}

type RepoParams struct {
	fx.In

	DB *gorm.DB
}

func NewRepo(p RepoParams) Repo {
	return &repoImpl{
		db: p.DB,
	}
}

func (r *repoImpl) FindByGuildID(ctx context.Context, guildID snowflake.ID) ([]AutoRole, error) {
	return gorm.G[AutoRole](r.db).Where("guild_id = ?", guildID).Find(ctx)
}

func (r *repoImpl) Save(ctx context.Context, autoRole *AutoRole) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "guild_id"}, {Name: "role_id"}},
			DoNothing: true,
		}).
		Create(autoRole).Error
}

func (r *repoImpl) DeleteByGuildIDAndRoleID(ctx context.Context, guildID, roleID snowflake.ID) (int, error) {
	return gorm.G[AutoRole](r.db).
		Where("guild_id = ? AND role_id = ?", guildID, roleID).
		Delete(ctx)
}
