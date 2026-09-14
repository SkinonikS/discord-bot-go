package tempvoicechannel

import (
	"uuid"

	"github.com/disgoorg/snowflake/v2"
)

type TempVoiceChannel struct {
	ID            uuid.UUID
	GuildID       snowflake.ID
	RootChannelID snowflake.ID
	ParentID      snowflake.ID
}
