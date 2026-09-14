package tempvoicechannelstate

import (
	"uuid"

	"github.com/disgoorg/snowflake/v2"
)

type TempVoiceChannelState struct {
	ID        uuid.UUID
	GuildID   snowflake.ID
	ChannelID snowflake.ID
}
