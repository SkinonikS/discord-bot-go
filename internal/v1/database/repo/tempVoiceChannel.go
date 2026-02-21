package repo

import (
	"context"
	"errors"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TempVoiceChannelRepo struct {
	db *gorm.DB
}

type TempVoiceChannelParams struct {
	fx.In
	DB *gorm.DB
}

func NewTempVoiceChannelRepo(p TempVoiceChannelParams) *TempVoiceChannelRepo {
	return &TempVoiceChannelRepo{
		db: p.DB,
	}
}

func (r *TempVoiceChannelRepo) FindByRootChannel(ctx context.Context, channelID string) (*model.TempVoiceChannel, error) {
	setup, err := gorm.G[*model.TempVoiceChannel](r.db).
		Where("root_channel_id = ?", channelID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return setup, nil
}

func (r *TempVoiceChannelRepo) Save(ctx context.Context, setup *model.TempVoiceChannel) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "root_channel_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"parent_id"}),
		}).
		Create(setup).Error
}

func (r *TempVoiceChannelRepo) DeleteByRootChannel(ctx context.Context, channelID string) error {
	_, err := gorm.G[*model.TempVoiceChannel](r.db).
		Where("root_channel_id = ?", channelID).
		Delete(ctx)
	return err
}
