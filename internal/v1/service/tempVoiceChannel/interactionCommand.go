package tempVoiceChannel

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/omit"
	"go.uber.org/fx"
)

type InteractionCommand struct {
	channelRepo *repo.TempVoiceChannelRepo
}

type InteractionCommandParams struct {
	fx.In
	ChannelRepo *repo.TempVoiceChannelRepo
}

func NewInteractionCommand(p InteractionCommandParams) *InteractionCommand {
	return &InteractionCommand{
		channelRepo: p.ChannelRepo,
	}
}

func (c *InteractionCommand) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	switch *data.SubCommandName {
	case "setup":
		return c.handleSetup(ctx, e)
	case "remove":
		return c.handleRemove(ctx, e)
	}

	return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
}

func (c *InteractionCommand) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		Description:              "Manage temporary voice channels",
		DefaultMemberPermissions: omit.NewPtr(disgodiscord.PermissionManageGuild),
		Contexts:                 []disgodiscord.InteractionContextType{disgodiscord.InteractionContextTypeGuild},
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "setup",
				Description: "Set up a temporary voice channel trigger",
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionChannel{
						Name:         "root_channel",
						Description:  "The voice channel that triggers temp channel creation",
						Required:     true,
						ChannelTypes: []disgodiscord.ChannelType{disgodiscord.ChannelTypeGuildVoice},
					},
					disgodiscord.ApplicationCommandOptionChannel{
						Name:         "parent_category",
						Description:  "The category where temp channels will be created",
						Required:     true,
						ChannelTypes: []disgodiscord.ChannelType{disgodiscord.ChannelTypeGuildCategory},
					},
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "remove",
				Description: "Remove a temporary voice channel trigger",
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionChannel{
						Name:         "root_channel",
						Description:  "The root voice channel to remove the trigger for",
						Required:     true,
						ChannelTypes: []disgodiscord.ChannelType{disgodiscord.ChannelTypeGuildVoice},
					},
				},
			},
		},
	}
}

func (c *InteractionCommand) Name() string {
	return "temp-voice"
}

func (c *InteractionCommand) handleSetup(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	rootChannelID := data.Channel("root_channel").ID
	parentID := data.Channel("parent_category").ID

	if err := c.channelRepo.Save(ctx, &model.TempVoiceChannel{
		RootChannelID: rootChannelID,
		ParentID:      parentID,
	}); err != nil {
		return fmt.Errorf("failed to save temp voice channel setup: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: fmt.Sprintf("Temporary voice channel setup saved! Root: <#%s>, Category: <#%s>", rootChannelID, parentID),
	}, disgorest.WithCtx(ctx))
}

func (c *InteractionCommand) handleRemove(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	rootChannelID := data.Channel("root_channel").ID

	if err := c.channelRepo.DeleteByRootChannel(ctx, rootChannelID); err != nil {
		return fmt.Errorf("failed to remove temp voice channel setup: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: fmt.Sprintf("Temporary voice channel trigger removed for <#%s>.", rootChannelID),
	}, disgorest.WithCtx(ctx))
}
