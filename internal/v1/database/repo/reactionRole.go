package repo

import (
	"context"
	"errors"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReactionRoleRepo struct {
	db *gorm.DB
}

type ReactionRoleParams struct {
	fx.In
	DB *gorm.DB
}

func NewReactionRoleRepo(p ReactionRoleParams) *ReactionRoleRepo {
	return &ReactionRoleRepo{
		db: p.DB,
	}
}

func (r *ReactionRoleRepo) FindByMessageAndEmoji(ctx context.Context, messageID, emojiName string) (*model.ReactionRole, error) {
	role, err := gorm.G[*model.ReactionRole](r.db).
		Where("message_id = ?", messageID).
		Where("emoji_name = ?", emojiName).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return role, nil
}

func (r *ReactionRoleRepo) Save(ctx context.Context, role *model.ReactionRole) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}, {Name: "emoji_name"}},
			DoUpdates: clause.AssignmentColumns([]string{"role_id"}),
		}).
		Create(role).Error
}

func (r *ReactionRoleRepo) DeleteByMessageAndEmoji(ctx context.Context, messageID, emojiName string) error {
	_, err := gorm.G[*model.ReactionRole](r.db).
		Where("message_id = ?", messageID).
		Where("emoji_name = ?", emojiName).
		Delete(ctx)
	return err
}
