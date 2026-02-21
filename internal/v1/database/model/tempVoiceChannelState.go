package model

import "github.com/google/uuid"

type TempVoiceChannelState struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	GuildID     string    `gorm:"not null;index"`
	ChannelID   string    `gorm:"not null;uniqueIndex"`
	MemberCount int       `gorm:"not null;default:0"`
}

func (*TempVoiceChannelState) TableName() string {
	return "temp_voice_channel_states"
}
