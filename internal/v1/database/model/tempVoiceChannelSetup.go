package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TempVoiceChannelSetup struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	GuildID       string    `gorm:"not null;uniqueIndex:idx_guild_root"`
	RootChannelID string    `gorm:"not null;uniqueIndex:idx_guild_root"`
	ParentID      string    `gorm:"not null"`
}

func (u *TempVoiceChannelSetup) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (*TempVoiceChannelSetup) TableName() string {
	return "temp_voice_channel_setups"
}
