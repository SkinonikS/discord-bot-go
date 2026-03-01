package tempVoiceChannel

import (
	"github.com/bwmarrin/discordgo"
)

func countChannelMembers(s *discordgo.Session, guildID, channelID string) int {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return -1
	}

	count := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == channelID {
			count++
		}
	}
	return count
}
