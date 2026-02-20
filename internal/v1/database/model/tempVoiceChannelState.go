package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TempVoiceChannelState struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	GuildID   string    `gorm:"not null;index"`
	UserID    string    `gorm:"not null"`
	ChannelID string    `gorm:"not null;uniqueIndex"`
}

func (u *TempVoiceChannelState) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (*TempVoiceChannelState) TableName() string {
	return "temp_voice_channel_states"
}
