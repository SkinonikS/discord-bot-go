package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/config"
)

const (
	ConfigKey = "discord"
)

var intentMap = map[string]discordgo.Intent{
	"guilds":                 discordgo.IntentsGuilds,
	"guildmembers":           discordgo.IntentsGuildMembers,
	"guildbans":              discordgo.IntentsGuildBans,
	"guildemojis":            discordgo.IntentsGuildEmojis,
	"guildintegrations":      discordgo.IntentsGuildIntegrations,
	"guildwebhooks":          discordgo.IntentsGuildWebhooks,
	"guildinvites":           discordgo.IntentsGuildInvites,
	"guildvoicestates":       discordgo.IntentsGuildVoiceStates,
	"guildpresences":         discordgo.IntentsGuildPresences,
	"guildmessages":          discordgo.IntentsGuildMessages,
	"guildmessagereactions":  discordgo.IntentsGuildMessageReactions,
	"guildmessagetyping":     discordgo.IntentsGuildMessageTyping,
	"directmessages":         discordgo.IntentsDirectMessages,
	"directmessagereactions": discordgo.IntentsDirectMessageReactions,
	"directmessagetyping":    discordgo.IntentsDirectMessageTyping,
	"messagecontent":         discordgo.IntentsMessageContent,
	"guildscheduledevents":   discordgo.IntentsGuildScheduledEvents,
	"allwithoutprivileged":   discordgo.IntentsAllWithoutPrivileged,
	"all":                    discordgo.IntentsAll,
	"none":                   discordgo.IntentsNone,
}

type Config struct {
	AppID   string   `yaml:"appID"`
	Token   string   `yaml:"token"`
	Intents []string `yaml:"intents"`
}

func (c *Config) ParseIntents() (discordgo.Intent, error) {
	if len(c.Intents) == 0 {
		return discordgo.IntentsNone, nil
	}

	var result discordgo.Intent
	for _, name := range c.Intents {
		intent, ok := intentMap[strings.ToLower(name)]
		if !ok {
			return 0, fmt.Errorf("unknown intent: %q", name)
		}
		result |= intent
	}
	return result, nil
}

func NewConfig(provider config.Provider) (*Config, error) {
	var cfg *Config
	if err := provider.Get(ConfigKey).Populate(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
