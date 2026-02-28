package reactionRole

import (
	"context"
	"fmt"

	"regexp"
	"strings"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	"github.com/SkinonikS/discord-bot-go/internal/v1/discord/permission"
	"github.com/bwmarrin/discordgo"
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

func (c *InteractionCommand) Execute(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate) error {
	if e.Type != discordgo.InteractionApplicationCommand {
		return nil
	}

	data := e.ApplicationCommandData()
	if len(data.Options) == 0 {
		return nil
	}

	subCmd := data.Options[0]
	switch subCmd.Name {
	case "add":
		return c.handleAdd(ctx, s, e, subCmd)
	case "remove":
		return c.handleRemove(ctx, s, e, subCmd)
	}
	return nil
}

func (c *InteractionCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:                     c.Name(),
		Description:              "Manage reaction roles",
		DefaultMemberPermissions: permission.New(discordgo.PermissionManageGuild).AsBitPtr(),
		Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a reaction role mapping",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "The channel where the message is located",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "message_id",
						Description: "The ID of the message to listen for reactions on",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "emoji",
						Description: "The emoji name that triggers role assignment",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "The role to assign when the emoji is reacted",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a reaction role mapping",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "message_id",
						Description: "The ID of the message",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "emoji",
						Description: "The emoji name to remove the mapping for",
						Required:    true,
					},
				},
			},
		},
	}
}

func (c *InteractionCommand) ForOwnerOnly() bool {
	return true
}

func (c *InteractionCommand) Name() string {
	return "reaction-role"
}

func (c *InteractionCommand) handleAdd(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate, subCmd *discordgo.ApplicationCommandInteractionDataOption) error {
	var channelID, messageID, emojiName, roleID string
	for _, opt := range subCmd.Options {
		switch opt.Name {
		case "channel":
			channelID = opt.ChannelValue(s).ID
		case "message_id":
			messageID = opt.StringValue()
		case "emoji":
			emojiName = opt.StringValue()
		case "role":
			roleID = opt.Value.(string)
		}
	}

	formattedEmoji := c.formatEmoji(emojiName)
	if err := c.repo.Save(ctx, &model.ReactionRole{
		ChannelID: channelID,
		MessageID: messageID,
		EmojiName: formattedEmoji,
		RoleID:    roleID,
	}); err != nil {
		return fmt.Errorf("failed to save reaction role: %w", err)
	}

	if err := s.MessageReactionAdd(channelID, messageID, formattedEmoji, discordgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("failed to add reaction to message: %w", err)
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Reaction role saved! Channel: <#%s>, Message: %s, Emoji: %s, Role: <@&%s>", channelID, messageID, emojiName, roleID),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *InteractionCommand) handleRemove(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate, subCmd *discordgo.ApplicationCommandInteractionDataOption) error {
	var messageID, emojiName string
	for _, opt := range subCmd.Options {
		switch opt.Name {
		case "message_id":
			messageID = opt.StringValue()
		case "emoji":
			emojiName = opt.StringValue()
		}
	}

	formattedEmoji := c.formatEmoji(emojiName)
	reaction, err := c.repo.FindByMessageAndEmoji(ctx, messageID, formattedEmoji)
	if err != nil {
		return fmt.Errorf("failed to find reaction role: %w", err)
	}

	if err := c.repo.DeleteByMessageAndEmoji(ctx, messageID, formattedEmoji); err != nil {
		return fmt.Errorf("failed to remove reaction role: %w", err)
	}

	if reaction != nil {
		if err := s.MessageReactionRemove(reaction.ChannelID, messageID, formattedEmoji, "@me", discordgo.WithContext(ctx)); err != nil {
			return fmt.Errorf("failed to remove reaction from message: %w", err)
		}
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Reaction role removed! Message: %s, Emoji: %s", messageID, emojiName),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *InteractionCommand) formatEmoji(emoji string) string {
	if matches := c.emojiRegex.FindStringSubmatch(emoji); len(matches) == 4 {
		return fmt.Sprintf("%s:%s", matches[2], matches[3])
	}
	return strings.TrimSpace(emoji)
}
