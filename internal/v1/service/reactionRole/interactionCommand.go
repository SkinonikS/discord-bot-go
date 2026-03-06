package reactionRole

import (
	"context"
	"fmt"

	"regexp"
	"strings"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/omit"
	"go.uber.org/fx"
)

type InteractionCommand struct {
	repo       *repo.ReactionRoleRepo
	emojiRegex *regexp.Regexp
}

type InteractionCommandParams struct {
	fx.In
	Repo *repo.ReactionRoleRepo
}

func NewInteractionCommand(p InteractionCommandParams) *InteractionCommand {
	return &InteractionCommand{
		repo:       p.Repo,
		emojiRegex: regexp.MustCompile(`^<(a?):(\w+):(\d+)>$`),
	}
}

func (c *InteractionCommand) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	switch *data.SubCommandName {
	case "add":
		return c.handleAdd(ctx, e)
	case "remove":
		return c.handleRemove(ctx, e)
	}

	return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
}

func (c *InteractionCommand) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		Description:              "Manage reaction roles",
		DefaultMemberPermissions: omit.NewPtr(disgodiscord.PermissionsNone.Add(disgodiscord.PermissionManageGuild)),
		Contexts:                 []disgodiscord.InteractionContextType{disgodiscord.InteractionContextTypeGuild},
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

func (c *InteractionCommand) Name() string {
	return "reaction-role"
}

func (c *InteractionCommand) handleAdd(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	channelID := data.Channel("channel").ID
	messageID := data.Snowflake("message_id")
	emojiName := data.String("emoji")
	role := data.Role("role")

	formattedEmoji := c.formatEmoji(emojiName)
	if err := c.repo.Save(ctx, &model.ReactionRole{
		ChannelID: channelID,
		MessageID: messageID,
		EmojiName: formattedEmoji,
		RoleID:    role.ID,
	}); err != nil {
		return fmt.Errorf("failed to save reaction role: %w", err)
	}

	if err := e.Client().Rest.AddReaction(channelID, messageID, formattedEmoji, disgorest.WithCtx(ctx)); err != nil {
		return fmt.Errorf("failed to add reaction to message: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: fmt.Sprintf("Reaction role saved! Channel: <#%s>, Message: %s, Emoji: %s, Role: <@&%s>", channelID, messageID, emojiName, role.ID),
	})
}

func (c *InteractionCommand) handleRemove(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	messageID := data.Snowflake("message_id")
	emojiName := data.String("emoji")

	formattedEmoji := c.formatEmoji(emojiName)
	reaction, err := c.repo.FindByMessageAndEmoji(ctx, messageID, formattedEmoji)
	if err != nil {
		return fmt.Errorf("failed to find reaction role: %w", err)
	}

	if err := c.repo.DeleteByMessageAndEmoji(ctx, messageID, formattedEmoji); err != nil {
		return fmt.Errorf("failed to remove reaction role: %w", err)
	}

	if reaction != nil {
		if err := e.Client().Rest.RemoveOwnReaction(reaction.ChannelID, messageID, formattedEmoji, disgorest.WithCtx(ctx)); err != nil {
			return fmt.Errorf("failed to remove reaction from message: %w", err)
		}
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: fmt.Sprintf("Reaction role removed! Message: %s, Emoji: %s", messageID, emojiName),
	})
}

func (c *InteractionCommand) formatEmoji(emoji string) string {
	if matches := c.emojiRegex.FindStringSubmatch(emoji); len(matches) == 4 {
		return fmt.Sprintf("%s:%s", matches[2], matches[3])
	}
	return strings.TrimSpace(emoji)
}
