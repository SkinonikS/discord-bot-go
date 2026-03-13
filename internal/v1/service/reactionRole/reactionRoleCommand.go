package reactionRole

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/SkinonikS/discord-bot-go/internal/v1/translator"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/fx"
)

const (
	ReactionRoleCommandName = "reaction-role"
)

type reactionRoleCommandImpl struct {
	t          translator.Translator
	service    Service
	emojiRegex *regexp.Regexp
}

type ReactionRoleCommandParams struct {
	fx.In

	T       translator.Translator
	Service Service
}

func NewReactionRoleCommand(p ReactionRoleCommandParams) interactionCommand.Command {
	return &reactionRoleCommandImpl{
		t:          p.T,
		service:    p.Service,
		emojiRegex: regexp.MustCompile(`^<(a?):(\w+):(\d+)>$`),
	}
}

func (c *reactionRoleCommandImpl) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	switch *data.SubCommandName {
	case "add":
		return c.handleAdd(ctx, e)
	case "remove":
		return c.handleRemove(ctx, e)
	}

	return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
}

func (c *reactionRoleCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		NameLocalizations:        c.t.SimpleLocalizeAll(c.Name()),
		Description:              "Manage reaction roles",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("Manage reaction roles"),
		DefaultMemberPermissions: omit.NewPtr(
			disgodiscord.PermissionsNone.Add(disgodiscord.PermissionManageGuild),
		),
		Contexts: []disgodiscord.InteractionContextType{
			disgodiscord.InteractionContextTypeGuild,
		},
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "add",
				NameLocalizations:        c.t.SimpleLocalizeAll("add"),
				Description:              "Add a reaction role mapping",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Add a reaction role mapping"),
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionChannel{
						Name:                     "channel",
						NameLocalizations:        c.t.SimpleLocalizeAll("channel"),
						Description:              "The channel where the message is located",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The channel where the message is located"),
						Required:                 true,
					},
					disgodiscord.ApplicationCommandOptionString{
						Name:                     "message_id",
						NameLocalizations:        c.t.SimpleLocalizeAll("message_id"),
						Description:              "The ID of the message to listen for reactions on",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The ID of the message to listen for reactions on"),
						Required:                 true,
					},
					disgodiscord.ApplicationCommandOptionString{
						Name:                     "emoji",
						NameLocalizations:        c.t.SimpleLocalizeAll("emoji"),
						Description:              "The emoji name that triggers role assignment",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The emoji name that triggers role assignment"),
						Required:                 true,
					},
					disgodiscord.ApplicationCommandOptionRole{
						Name:                     "role",
						NameLocalizations:        c.t.SimpleLocalizeAll("role"),
						Description:              "The role to assign when the emoji is reacted",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The role to assign when the emoji is reacted"),
						Required:                 true,
					},
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "remove",
				NameLocalizations:        c.t.SimpleLocalizeAll("remove"),
				Description:              "Remove a reaction role mapping",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Remove a reaction role mapping"),
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionString{
						Name:                     "message_id",
						NameLocalizations:        c.t.SimpleLocalizeAll("message_id"),
						Description:              "The ID of the message",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The ID of the message"),
						Required:                 true,
					},
					disgodiscord.ApplicationCommandOptionString{
						Name:                     "emoji",
						NameLocalizations:        c.t.SimpleLocalizeAll("emoji"),
						Description:              "The emoji name to remove the mapping for",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The emoji name to remove the mapping for"),
						Required:                 true,
					},
				},
			},
		},
	}
}

func (c *reactionRoleCommandImpl) Name() string {
	return ReactionRoleCommandName
}

func (c *reactionRoleCommandImpl) handleAdd(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	channelID := data.Channel("channel").ID
	messageID := data.Snowflake("message_id")
	emojiName := data.String("emoji")
	role := data.Role("role")

	if _, err := c.service.CreateReactionRole(ctx, CreateReactionRole{
		GuildID:   *e.GuildID(),
		ChannelID: channelID,
		MessageID: messageID,
		EmojiName: emojiName,
		RoleID:    role.ID,
	}); err != nil {
		return fmt.Errorf("failed to save reaction role: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "Reaction role saved! Channel: <#{{.ChannelID}}>, Message: {{.MessageID}}, Emoji: {{.Emoji}}, Role: <@&{{.RoleID}}>",
			TemplateData: map[string]any{
				"ChannelID": channelID,
				"MessageID": messageID,
				"Emoji":     emojiName,
				"RoleID":    role.ID,
			},
		}),
	})
}

func (c *reactionRoleCommandImpl) handleRemove(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	messageID := data.Snowflake("message_id")
	emojiName := data.String("emoji")

	if err := c.service.DeleteReactionRole(ctx, DeleteReactionRole{
		GuildID:   *e.GuildID(),
		MessageID: messageID,
		EmojiName: emojiName,
	}); err != nil {
		if errors.Is(err, ErrReactionRoleNotFound) {
			return e.CreateMessage(disgodiscord.MessageCreate{
				Flags:   disgodiscord.MessageFlagEphemeral,
				Content: c.t.SimpleLocalize(e.Locale(), "No reaction role found for this message and emoji."),
			})
		}

		return fmt.Errorf("failed to remove reaction role: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "Reaction role removed! Message: {{.MessageID}}, Emoji: {{.Emoji}}",
			TemplateData: map[string]any{
				"MessageID": messageID,
				"Emoji":     emojiName,
			},
		}),
	})
}
