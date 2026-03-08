package discord

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/config"
)

const (
	ConfigKey = "discord"
)

var intentMap = map[string]gateway.Intents{
	"guilds":                      gateway.IntentGuilds,
	"guildmembers":                gateway.IntentGuildMembers,
	"guildmoderation":             gateway.IntentGuildModeration,
	"guildexpressions":            gateway.IntentGuildExpressions,
	"guildintegrations":           gateway.IntentGuildIntegrations,
	"guildwebhooks":               gateway.IntentGuildWebhooks,
	"guildinvites":                gateway.IntentGuildInvites,
	"guildvoicestates":            gateway.IntentGuildVoiceStates,
	"guildpresences":              gateway.IntentGuildPresences,
	"guildmessages":               gateway.IntentGuildMessages,
	"guildmessagereactions":       gateway.IntentGuildMessageReactions,
	"guildmessagetyping":          gateway.IntentGuildMessageTyping,
	"directmessages":              gateway.IntentDirectMessages,
	"directmessagereactions":      gateway.IntentDirectMessageReactions,
	"directmessagetyping":         gateway.IntentDirectMessageTyping,
	"messagecontent":              gateway.IntentMessageContent,
	"guildscheduledevents":        gateway.IntentGuildScheduledEvents,
	"automoderationconfiguration": gateway.IntentAutoModerationConfiguration,
	"automoderationexecution":     gateway.IntentAutoModerationExecution,
	"guildmessagepolls":           gateway.IntentGuildMessagePolls,
	"directmessagepolls":          gateway.IntentDirectMessagePolls,
	"all":                         gateway.IntentsAll,
	"none":                        gateway.IntentsNone,
}

type Config struct {
	AppID       snowflake.ID `yaml:"appID"`
	Token       string       `yaml:"token"`
	Intents     []string     `yaml:"intents"`
	WorkerCount uint16       `yaml:"workerCount"`
}

func (c *Config) ParseIntents() (gateway.Intents, error) {
	if len(c.Intents) == 0 {
		return gateway.IntentsNone, nil
	}

	var result gateway.Intents
	for _, name := range c.Intents {
		intent, ok := intentMap[strings.ToLower(name)]
		if !ok {
			return 0, fmt.Errorf("unknown intent: %q", name)
		}
		result = result.Add(intent)
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
