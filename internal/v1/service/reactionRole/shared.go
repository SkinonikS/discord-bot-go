package reactionRole

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func emojiKey(emoji discordgo.Emoji) string {
	if emoji.ID != "" {
		return fmt.Sprintf("%s:%s", emoji.Name, emoji.ID)
	}
	return emoji.Name
}
