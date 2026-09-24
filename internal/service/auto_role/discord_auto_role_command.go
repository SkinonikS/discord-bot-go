package autorole

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/SkinonikS/discord-bot-go/internal/infra/translator"
	interactioncommand "github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	autorole "github.com/SkinonikS/discord-bot-go/internal/service/repository/auto_role"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

type discordAutoRoleCommandImpl struct {
	t       translator.Translator
	service Service
}

type DiscordAutoRoleCommandParams struct {
	fx.In

	T       translator.Translator
	Service Service
}

func NewDiscordAutoRoleCommand(p DiscordAutoRoleCommandParams) interactioncommand.Command {
	return &discordAutoRoleCommandImpl{
		t:       p.T,
		service: p.Service,
	}
}

func (c *discordAutoRoleCommandImpl) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	switch *data.SubCommandName {
	case "add":
		return c.handleAdd(ctx, e)
	case "remove":
		return c.handleRemove(ctx, e)
	case "list":
		return c.handleList(ctx, e)
	}

	return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
}

func (c *discordAutoRoleCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		NameLocalizations:        c.t.SimpleLocalizeAll(c.Name()),
		Description:              "Manage roles automatically assigned to new members",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("Manage roles automatically assigned to new members"),
		DefaultMemberPermissions: omit.NewPtr(
			disgodiscord.PermissionsNone.Add(disgodiscord.PermissionManageRoles),
		),
		Contexts: []disgodiscord.InteractionContextType{
			disgodiscord.InteractionContextTypeGuild,
		},
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "add",
				NameLocalizations:        c.t.SimpleLocalizeAll("add"),
				Description:              "Assign a role automatically when a member joins",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Assign a role automatically when a member joins"),
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionRole{
						Name:                     "role",
						NameLocalizations:        c.t.SimpleLocalizeAll("role"),
						Description:              "The role to assign to new members",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The role to assign to new members"),
						Required:                 true,
					},
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "remove",
				NameLocalizations:        c.t.SimpleLocalizeAll("remove"),
				Description:              "Stop automatically assigning a role to new members",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Stop automatically assigning a role to new members"),
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionRole{
						Name:                     "role",
						NameLocalizations:        c.t.SimpleLocalizeAll("role"),
						Description:              "The role to stop assigning",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The role to stop assigning"),
						Required:                 true,
					},
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "list",
				NameLocalizations:        c.t.SimpleLocalizeAll("list"),
				Description:              "List roles automatically assigned to new members",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("List roles automatically assigned to new members"),
			},
		},
	}
}

func (c *discordAutoRoleCommandImpl) Name() string {
	return "auto-role"
}

func (c *discordAutoRoleCommandImpl) Scope() interactioncommand.CommandScope {
	return interactioncommand.CommandScopeGuild
}

func (c *discordAutoRoleCommandImpl) handleAdd(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	role := data.Role("role")

	if _, err := c.service.AddAutoRole(ctx, AddAutoRoleParams{
		GuildID: *e.GuildID(),
		RoleID:  role.ID,
	}); err != nil {
		if errors.Is(err, ErrAutoRoleAlreadyExists) {
			return e.CreateMessage(disgodiscord.MessageCreate{
				Flags: disgodiscord.MessageFlagEphemeral,
				Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
					MessageID: "<@&{{.RoleID}}> is already assigned automatically.",
					TemplateData: map[string]any{
						"RoleID": role.ID,
					},
				}),
			})
		}

		return fmt.Errorf("failed to save auto role: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "<@&{{.RoleID}}> will now be assigned automatically to new members.",
			TemplateData: map[string]any{
				"RoleID": role.ID,
			},
		}),
	})
}

func (c *discordAutoRoleCommandImpl) handleRemove(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	role := data.Role("role")

	if err := c.service.RemoveAutoRole(ctx, RemoveAutoRoleParams{
		GuildID: *e.GuildID(),
		RoleID:  role.ID,
	}); err != nil {
		if errors.Is(err, ErrAutoRoleNotFound) {
			return e.CreateMessage(disgodiscord.MessageCreate{
				Flags: disgodiscord.MessageFlagEphemeral,
				Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
					MessageID: "<@&{{.RoleID}}> is not assigned automatically.",
					TemplateData: map[string]any{
						"RoleID": role.ID,
					},
				}),
			})
		}

		return fmt.Errorf("failed to remove auto role: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "<@&{{.RoleID}}> will no longer be assigned automatically.",
			TemplateData: map[string]any{
				"RoleID": role.ID,
			},
		}),
	})
}

func (c *discordAutoRoleCommandImpl) handleList(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	autoRoles, err := c.service.ListAutoRoles(ctx, ListAutoRolesParams{
		GuildID: *e.GuildID(),
	})
	if err != nil {
		return fmt.Errorf("failed to list auto roles: %w", err)
	}

	if len(autoRoles) == 0 {
		return e.CreateMessage(disgodiscord.MessageCreate{
			Flags:   disgodiscord.MessageFlagEphemeral,
			Content: c.t.SimpleLocalize(e.Locale(), "No roles are currently assigned automatically."),
		})
	}

	mentions := lo.Map(autoRoles, func(autoRole autorole.AutoRole, _ int) string {
		return fmt.Sprintf("<@&%d>", autoRole.RoleID)
	})

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "Roles assigned automatically to new members: {{.Roles}}",
			TemplateData: map[string]any{
				"Roles": strings.Join(mentions, ", "),
			},
		}),
	})
}
