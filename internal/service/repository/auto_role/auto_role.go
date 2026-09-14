package autorole

import (
	"uuid"

	"github.com/disgoorg/snowflake/v2"
)

type AutoRole struct {
	ID      uuid.UUID
	GuildID snowflake.ID
	RoleID  snowflake.ID
}
