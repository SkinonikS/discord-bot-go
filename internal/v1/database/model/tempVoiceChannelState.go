package model

import "github.com/google/uuid"

type TempVoiceChannelState struct {
	ID          uuid.UUID
	ChannelID   string
	MemberCount int
}

func (*TempVoiceChannelState) TableName() string {
	return "temp_voice_channel_states"
}
