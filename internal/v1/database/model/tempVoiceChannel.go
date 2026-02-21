package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TempVoiceChannel struct {
	ID            uuid.UUID
	RootChannelID string
	ParentID      string
}

func (u *TempVoiceChannel) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (*TempVoiceChannel) TableName() string {
	return "temp_voice_channels"
}
