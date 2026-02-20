package repo

import (
	"context"
	"errors"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TempVoiceChannelSetupRepo struct {
	db *gorm.DB
}

type TempVoiceChannelSetupParams struct {
	fx.In
	DB *gorm.DB
}

func NewTempVoiceChannelSetupRepo(p TempVoiceChannelSetupParams) *TempVoiceChannelSetupRepo {
	return &TempVoiceChannelSetupRepo{
		db: p.DB,
	}
}

func (r *TempVoiceChannelSetupRepo) FindByRootChannel(ctx context.Context, guildID, channelID string) (*model.TempVoiceChannelSetup, error) {
	setup, err := gorm.G[*model.TempVoiceChannelSetup](r.db).
		Where("guild_id = ? AND root_channel_id = ?", guildID, channelID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return setup, nil
}

func (r *TempVoiceChannelSetupRepo) Save(ctx context.Context, setup *model.TempVoiceChannelSetup) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "guild_id"}, {Name: "root_channel_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"parent_id"}),
		}).
		Create(setup).Error
}

func (r *TempVoiceChannelSetupRepo) DeleteByRootChannel(ctx context.Context, guildID, channelID string) error {
	_, err := gorm.G[*model.TempVoiceChannelSetup](r.db).
		Where("guild_id = ? AND root_channel_id = ?", guildID, channelID).
		Delete(ctx)
	return err
}
