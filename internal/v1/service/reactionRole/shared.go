package reactionRole

import (
	"fmt"

	disgodiscord "github.com/disgoorg/disgo/discord"
)

func emojiKey(emoji disgodiscord.PartialEmoji) string {
	if emoji.ID != nil && emoji.Name != nil {
		return fmt.Sprintf("%s:%s", emoji.Name, emoji.ID.String())
	}
	return emoji.Reaction()
}
