package guildcommandsetting

import (
	"uuid"

	"github.com/disgoorg/snowflake/v2"
)

type DisabledGuildCommand struct {
	ID          uuid.UUID
	GuildID     snowflake.ID
	CommandName string
}
