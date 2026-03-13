package tempVoiceChannel

import (
	"context"
	"errors"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/SkinonikS/discord-bot-go/internal/v1/translator"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/omit"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/fx"
)

const (
	TempVoiceCommandName = "temp-voice"
)

type tempVoiceCommandImpl struct {
	t       translator.Translator
	service Service
}

type TempVoiceCommandParams struct {
	fx.In

	T       translator.Translator
	Service Service
}

func NewTempVoiceCommand(p TempVoiceCommandParams) interactionCommand.Command {
	return &tempVoiceCommandImpl{
		t:       p.T,
		service: p.Service,
	}
}

func (c *tempVoiceCommandImpl) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	switch *data.SubCommandName {
	case "setup":
		return c.handleSetup(ctx, e)
	case "remove":
		return c.handleRemove(ctx, e)
	}

	return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
}

func (c *tempVoiceCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		NameLocalizations:        c.t.SimpleLocalizeAll(c.Name()),
		Description:              "Manage temporary voice channels",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("Manage temporary voice channels"),
		DefaultMemberPermissions: omit.NewPtr(disgodiscord.PermissionManageGuild),
		Contexts: []disgodiscord.InteractionContextType{
			disgodiscord.InteractionContextTypeGuild,
		},
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "setup",
				NameLocalizations:        c.t.SimpleLocalizeAll("setup"),
				Description:              "Set up a temporary voice channel trigger",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Set up a temporary voice channel trigger"),
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionChannel{
						Name:                     "root_channel",
						NameLocalizations:        c.t.SimpleLocalizeAll("root_channel"),
						Description:              "The voice channel that triggers temp channel creation",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The voice channel that triggers temp channel creation"),
						Required:                 true,
						ChannelTypes: []disgodiscord.ChannelType{
							disgodiscord.ChannelTypeGuildVoice,
						},
					},
					disgodiscord.ApplicationCommandOptionChannel{
						Name:                     "parent_category",
						NameLocalizations:        c.t.SimpleLocalizeAll("parent_category"),
						Description:              "The category where temp channels will be created",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The category where temp channels will be created"),
						Required:                 true,
						ChannelTypes: []disgodiscord.ChannelType{
							disgodiscord.ChannelTypeGuildCategory,
						},
					},
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "remove",
				NameLocalizations:        c.t.SimpleLocalizeAll("remove"),
				Description:              "Remove a temporary voice channel trigger",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Remove a temporary voice channel trigger"),
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionChannel{
						Name:                     "root_channel",
						NameLocalizations:        c.t.SimpleLocalizeAll("root_channel"),
						Description:              "The root voice channel to remove the trigger for",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The root voice channel to remove the trigger for"),
						Required:                 true,
						ChannelTypes: []disgodiscord.ChannelType{
							disgodiscord.ChannelTypeGuildVoice,
						},
					},
				},
			},
		},
	}
}

func (c *tempVoiceCommandImpl) Name() string {
	return TempVoiceCommandName
}

func (c *tempVoiceCommandImpl) handleSetup(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
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
		Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "Temporary voice channel setup saved! Root: <#{{.RootChannelID}}>, Category: <#{{.ParentCategoryID}}>",
			TemplateData: map[string]any{
				"RootChannelID":    rootChannelID,
				"ParentCategoryID": parentCategoryID,
			},
		}),
	}, disgorest.WithCtx(ctx))
}

func (c *tempVoiceCommandImpl) handleRemove(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	rootChannelID := data.Channel("root_channel").ID

	if err := c.service.DeleteSetupChannel(ctx, DeleteSetupChannel{
		GuildID:       *e.GuildID(),
		RootChannelID: rootChannelID,
	}); err != nil {
		if errors.Is(err, ErrSetupChannelNotFound) {
			return e.CreateMessage(disgodiscord.MessageCreate{
				Flags:   disgodiscord.MessageFlagEphemeral,
				Content: c.t.SimpleLocalize(e.Locale(), "No temp voice channel setup found for this channel."),
			}, disgorest.WithCtx(ctx))
		}

		return fmt.Errorf("failed to delete temp voice channel setup: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "Temporary voice channel trigger removed for <#{{.RootChannelID}}>",
			TemplateData: map[string]any{
				"RootChannelID": rootChannelID,
			},
		}),
	}, disgorest.WithCtx(ctx))
}
