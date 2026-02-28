package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TempVoiceChannelState struct {
	ID          uuid.UUID
	ChannelID   string
	MemberCount int
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
