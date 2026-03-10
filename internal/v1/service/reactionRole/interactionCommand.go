package reactionRole

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"
	"go.uber.org/fx"
)

const (
	InteractionCommandName = "reaction-role"
)

type interactionCommandImpl struct {
	service    Service
	emojiRegex *regexp.Regexp
}

type InteractionCommandParams struct {
	fx.In

	Service Service
}

func NewInteractionCommand(p InteractionCommandParams) interactionCommand.Command {
	return &interactionCommandImpl{
		service:    p.Service,
		emojiRegex: regexp.MustCompile(`^<(a?):(\w+):(\d+)>$`),
	}
}

func (c *interactionCommandImpl) Execute(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	data := e.SlashCommandInteractionData()

	switch *data.SubCommandName {
	case "add":
		return c.handleAdd(ctx, e)
	case "remove":
		return c.handleRemove(ctx, e)
	}

	return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
}

func (c *interactionCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:        c.Name(),
		Description: "Manage reaction roles",
		DefaultMemberPermissions: omit.NewPtr(
			disgodiscord.PermissionsNone.Add(disgodiscord.PermissionManageGuild),
		),
		Contexts: []disgodiscord.InteractionContextType{
			disgodiscord.InteractionContextTypeGuild,
		},
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "add",
				Description: "Add a reaction role mapping",
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: "The channel where the message is located",
						Required:    true,
					},
					disgodiscord.ApplicationCommandOptionString{
						Name:        "message_id",
						Description: "The ID of the message to listen for reactions on",
						Required:    true,
					},
					disgodiscord.ApplicationCommandOptionString{
						Name:        "emoji",
						Description: "The emoji name that triggers role assignment",
						Required:    true,
					},
					disgodiscord.ApplicationCommandOptionRole{
						Name:        "role",
						Description: "The role to assign when the emoji is reacted",
						Required:    true,
					},
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "remove",
				Description: "Remove a reaction role mapping",
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionString{
						Name:        "message_id",
						Description: "The ID of the message",
						Required:    true,
					},
					disgodiscord.ApplicationCommandOptionString{
						Name:        "emoji",
						Description: "The emoji name to remove the mapping for",
						Required:    true,
					},
				},
			},
		},
	}
}

func (c *interactionCommandImpl) Name() string {
	return InteractionCommandName
}

func (c *interactionCommandImpl) handleAdd(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
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
		Content: fmt.Sprintf(
			"Reaction role saved! Channel: <#%s>, Message: %s, Emoji: %s, Role: <@&%s>",
			channelID,
			messageID,
			emojiName,
			role.ID,
		),
	})
}

func (c *interactionCommandImpl) handleRemove(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
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
				Content: "No reaction role found for this message and emoji.",
			})
		}

		return fmt.Errorf("failed to remove reaction role: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: fmt.Sprintf("Reaction role removed! Message: %s, Emoji: %s", messageID, emojiName),
	})
}
