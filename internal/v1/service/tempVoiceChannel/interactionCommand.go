package tempVoiceChannel

import (
	"context"
	"errors"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/omit"
	"go.uber.org/fx"
)

const (
	InteractionCommandName = "temp-voice"
)

type interactionCommandImpl struct {
	service Service
}

type InteractionCommandParams struct {
	fx.In

	Service Service
}

func NewInteractionCommand(p InteractionCommandParams) interactionCommand.Command {
	return &interactionCommandImpl{
		service: p.Service,
	}
}

func (c *interactionCommandImpl) Execute(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	data := e.SlashCommandInteractionData()

	switch *data.SubCommandName {
	case "setup":
		return c.handleSetup(ctx, e)
	case "remove":
		return c.handleRemove(ctx, e)
	}

	return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
}

func (c *interactionCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		Description:              "Manage temporary voice channels",
		DefaultMemberPermissions: omit.NewPtr(disgodiscord.PermissionManageGuild),
		Contexts: []disgodiscord.InteractionContextType{
			disgodiscord.InteractionContextTypeGuild,
		},
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "setup",
				Description: "Set up a temporary voice channel trigger",
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionChannel{
						Name:        "root_channel",
						Description: "The voice channel that triggers temp channel creation",
						Required:    true,
						ChannelTypes: []disgodiscord.ChannelType{
							disgodiscord.ChannelTypeGuildVoice,
						},
					},
					disgodiscord.ApplicationCommandOptionChannel{
						Name:        "parent_category",
						Description: "The category where temp channels will be created",
						Required:    true,
						ChannelTypes: []disgodiscord.ChannelType{
							disgodiscord.ChannelTypeGuildCategory,
						},
					},
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "remove",
				Description: "Remove a temporary voice channel trigger",
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionChannel{
						Name:        "root_channel",
						Description: "The root voice channel to remove the trigger for",
						Required:    true,
						ChannelTypes: []disgodiscord.ChannelType{
							disgodiscord.ChannelTypeGuildVoice,
						},
					},
				},
			},
		},
	}
}

func (c *interactionCommandImpl) Name() string {
	return InteractionCommandName
}

func (c *interactionCommandImpl) handleSetup(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	data := e.SlashCommandInteractionData()
	rootChannelID := data.Channel("root_channel").ID
	parentCategoryID := data.Channel("parent_category").ID

	if _, err := c.service.CreateSetupChannel(ctx, SetupChannel{
		GuildID:          *e.GuildID(),
		RootChannelID:    rootChannelID,
		ParentCategoryID: parentCategoryID,
	}); err != nil {
		return fmt.Errorf("failed to save temp voice channel setup: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Content: fmt.Sprintf(
			"Temporary voice channel setup saved! Root: <#%s>, Category: <#%s>",
			rootChannelID,
			parentCategoryID,
		),
	}, disgorest.WithCtx(ctx))
}

func (c *interactionCommandImpl) handleRemove(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	data := e.SlashCommandInteractionData()
	rootChannelID := data.Channel("root_channel").ID

	if err := c.service.DeleteSetupChannel(ctx, DeleteSetupChannel{
		GuildID:   *e.GuildID(),
		ChannelID: rootChannelID,
	}); err != nil {
		if errors.Is(err, ErrSetupChannelNotFound) {
			return e.CreateMessage(disgodiscord.MessageCreate{
				Flags:   disgodiscord.MessageFlagEphemeral,
				Content: "No temp voice channel setup found for this root channel.",
			}, disgorest.WithCtx(ctx))
		}

		return fmt.Errorf("failed to delete temp voice channel setup: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: fmt.Sprintf("Temporary voice channel trigger removed for <#%s>.", rootChannelID),
	}, disgorest.WithCtx(ctx))
}
