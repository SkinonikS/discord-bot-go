package tempVoiceChannel

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
)

type InteractionCommand struct {
	repo *repo.TempVoiceChannelSetupRepo
}

type InteractionCommandParams struct {
	fx.In
	Repo *repo.TempVoiceChannelSetupRepo
}

func NewInteractionCommand(p InteractionCommandParams) *InteractionCommand {
	return &InteractionCommand{
		repo: p.Repo,
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
	case "setup":
		return c.handleSetup(ctx, s, e, subCmd)
	case "remove":
		return c.handleRemove(ctx, s, e, subCmd)
	}
	return nil
}

func (c *InteractionCommand) Definition() *discordgo.ApplicationCommand {
	var manageServerPerm int64 = discordgo.PermissionManageGuild
	return &discordgo.ApplicationCommand{
		Name:                     c.Name(),
		Description:              "Manage temporary voice channels",
		DefaultMemberPermissions: &manageServerPerm,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "setup",
				Description: "Set up a temporary voice channel trigger",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionChannel,
						Name:         "root_channel",
						Description:  "The voice channel that triggers temp channel creation",
						Required:     true,
						ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildVoice},
					},
					{
						Type:         discordgo.ApplicationCommandOptionChannel,
						Name:         "parent_category",
						Description:  "The category where temp channels will be created",
						Required:     true,
						ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildCategory},
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a temporary voice channel trigger",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionChannel,
						Name:         "root_channel",
						Description:  "The root voice channel to remove the trigger for",
						Required:     true,
						ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildVoice},
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
	return "temp-voice"
}

func (c *InteractionCommand) handleSetup(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate, subCmd *discordgo.ApplicationCommandInteractionDataOption) error {
	var rootChannelID, parentID string
	for _, opt := range subCmd.Options {
		switch opt.Name {
		case "root_channel":
			rootChannelID = opt.Value.(string)
		case "parent_category":
			parentID = opt.Value.(string)
		}
	}

	if err := c.repo.Save(ctx, &model.TempVoiceChannelSetup{
		GuildID:       e.GuildID,
		RootChannelID: rootChannelID,
		ParentID:      parentID,
	}); err != nil {
		return fmt.Errorf("failed to save temp voice channel setup: %w", err)
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Temporary voice channel setup saved! Root: <#%s>, Category: <#%s>", rootChannelID, parentID),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx))
}

func (c *InteractionCommand) handleRemove(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate, subCmd *discordgo.ApplicationCommandInteractionDataOption) error {
	var rootChannelID string
	for _, opt := range subCmd.Options {
		if opt.Name == "root_channel" {
			rootChannelID = opt.Value.(string)
		}
	}

	if err := c.repo.DeleteByRootChannel(ctx, e.GuildID, rootChannelID); err != nil {
		return fmt.Errorf("failed to remove temp voice channel setup: %w", err)
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Temporary voice channel trigger removed for <#%s>.", rootChannelID),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx))
}
