package repo

import (
	"context"
	"errors"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type TempVoiceChannelStateRepo struct {
	db *gorm.DB
}

type TempVoiceChannelStateParams struct {
	fx.In
	DB *gorm.DB
}

func NewTempVoiceChannelStateRepo(p TempVoiceChannelStateParams) *TempVoiceChannelStateRepo {
	return &TempVoiceChannelStateRepo{db: p.DB}
}

func (r *TempVoiceChannelStateRepo) FindByChannelID(ctx context.Context, channelID snowflake.ID) (*model.TempVoiceChannelState, error) {
	state, err := gorm.G[*model.TempVoiceChannelState](r.db).
		Where("channel_id = ?", channelID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return state, nil
}

func (r *TempVoiceChannelStateRepo) Save(ctx context.Context, state *model.TempVoiceChannelState) error {
	return r.db.WithContext(ctx).Create(state).Error
}

func (r *TempVoiceChannelStateRepo) DeleteByChannelID(ctx context.Context, channelID snowflake.ID) error {
	_, err := gorm.G[*model.TempVoiceChannelState](r.db).
		Where("channel_id = ?", channelID).
		Delete(ctx)
	return err
}

func (r *TempVoiceChannelStateRepo) FindAll(ctx context.Context) ([]*model.TempVoiceChannelState, error) {
	return gorm.G[*model.TempVoiceChannelState](r.db).Find(ctx)
}
