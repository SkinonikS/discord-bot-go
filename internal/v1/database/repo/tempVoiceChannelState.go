package repo

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/google/uuid"
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

func (r *TempVoiceChannelStateRepo) Save(ctx context.Context, state *model.TempVoiceChannelState) error {
	state.ID = uuid.New()
	return r.db.WithContext(ctx).Create(state).Error
}

func (r *TempVoiceChannelStateRepo) FindAllByGuild(ctx context.Context, guildID string) ([]*model.TempVoiceChannelState, error) {
	states, err := gorm.G[*model.TempVoiceChannelState](r.db).
		Where("guild_id = ?", guildID).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return states, nil
}

func (r *TempVoiceChannelStateRepo) DeleteByChannelID(ctx context.Context, channelID string) error {
	_, err := gorm.G[*model.TempVoiceChannelState](r.db).
		Where("channel_id = ?", channelID).
		Delete(ctx)
	return err
}
