package reactionrole

import (
	"uuid"

	"github.com/disgoorg/snowflake/v2"
)

type ReactionRole struct {
	ID        uuid.UUID
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	MessageID snowflake.ID
	EmojiName string
	RoleID    snowflake.ID
}
