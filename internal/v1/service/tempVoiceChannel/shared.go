package tempVoiceChannel

import (
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
)

func countChannelMembers(client *disgobot.Client, guildID, channelID snowflake.ID) int {
	count := 0
	for vs := range client.Caches.VoiceStates(guildID) {
		if vs.ChannelID == nil {
			continue
		}

		if *vs.ChannelID == channelID {
			count++
		}
	}

	return count
}
