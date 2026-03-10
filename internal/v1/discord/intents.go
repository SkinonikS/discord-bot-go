package discord

import (
	"fmt"
	"strings"

	disgogateway "github.com/disgoorg/disgo/gateway"
)

func intentMap() map[string]disgogateway.Intents {
	return map[string]disgogateway.Intents{
		"guilds":                      disgogateway.IntentGuilds,
		"guildmembers":                disgogateway.IntentGuildMembers,
		"guildmoderation":             disgogateway.IntentGuildModeration,
		"guildexpressions":            disgogateway.IntentGuildExpressions,
		"guildintegrations":           disgogateway.IntentGuildIntegrations,
		"guildwebhooks":               disgogateway.IntentGuildWebhooks,
		"guildinvites":                disgogateway.IntentGuildInvites,
		"guildvoicestates":            disgogateway.IntentGuildVoiceStates,
		"guildpresences":              disgogateway.IntentGuildPresences,
		"guildmessages":               disgogateway.IntentGuildMessages,
		"guildmessagereactions":       disgogateway.IntentGuildMessageReactions,
		"guildmessagetyping":          disgogateway.IntentGuildMessageTyping,
		"directmessages":              disgogateway.IntentDirectMessages,
		"directmessagereactions":      disgogateway.IntentDirectMessageReactions,
		"directmessagetyping":         disgogateway.IntentDirectMessageTyping,
		"messagecontent":              disgogateway.IntentMessageContent,
		"guildscheduledevents":        disgogateway.IntentGuildScheduledEvents,
		"automoderationconfiguration": disgogateway.IntentAutoModerationConfiguration,
		"automoderationexecution":     disgogateway.IntentAutoModerationExecution,
		"guildmessagepolls":           disgogateway.IntentGuildMessagePolls,
		"directmessagepolls":          disgogateway.IntentDirectMessagePolls,
		"all":                         disgogateway.IntentsAll,
		"none":                        disgogateway.IntentsNone,
	}
}

type Intents disgogateway.Intents

func (i *Intents) UnmarshalYAML(unmarshal func(any) error) error {
	var names []string
	if err := unmarshal(&names); err != nil {
		return err
	}

	availableIntents := intentMap()
	var result disgogateway.Intents
	for _, name := range names {
		intent, ok := availableIntents[strings.ToLower(name)]
		if !ok {
			return fmt.Errorf("unknown intent: %q", name)
		}
		result = result.Add(intent)
	}

	*i = Intents(result)
	return nil
}

func (i Intents) MarshalYAML() (any, error) {
	availableIntents := intentMap()
	var names []string
	for name, intent := range availableIntents {
		if name == "all" || name == "none" {
			continue
		}
		if disgogateway.Intents(i).Has(intent) {
			names = append(names, name)
		}
	}
	return names, nil
}

func (i Intents) Gateway() disgogateway.Intents {
	return disgogateway.Intents(i)
}
