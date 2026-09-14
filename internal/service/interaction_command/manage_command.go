package interactioncommand

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/SkinonikS/discord-bot-go/internal/infra/translator"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

const (
	ManageCommandName = "commands"
)

type manageCommandImpl struct {
	t        translator.Translator
	registry Registry
	settings Settings
	syncer   Syncer
}

type ManageCommandParams struct {
	fx.In

	T        translator.Translator
	Registry Registry
	Settings Settings
	Syncer   Syncer
}

func NewManageCommand(p ManageCommandParams) Command {
	return &manageCommandImpl{
		t:        p.T,
		registry: p.Registry,
		settings: p.Settings,
		syncer:   p.Syncer,
	}
}

func (c *manageCommandImpl) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	switch *data.SubCommandName {
	case "list":
		return c.handleList(ctx, e)
	case "enable":
		return c.handleSetDisabled(ctx, e, false)
	case "disable":
		return c.handleSetDisabled(ctx, e, true)
	}

	return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
}

func (c *manageCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	commandOption := disgodiscord.ApplicationCommandOptionString{
		Name:                     "command",
		NameLocalizations:        c.t.SimpleLocalizeAll("command"),
		Description:              "The command",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("The command"),
		Required:                 true,
		Choices:                  c.commandChoices(),
	}

	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		NameLocalizations:        c.t.SimpleLocalizeAll(c.Name()),
		Description:              "Manage which commands are enabled in this server",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("Manage which commands are enabled in this server"),
		DefaultMemberPermissions: omit.NewPtr(
			disgodiscord.PermissionsNone.Add(disgodiscord.PermissionAdministrator),
		),
		Contexts: []disgodiscord.InteractionContextType{
			disgodiscord.InteractionContextTypeGuild,
		},
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "list",
				NameLocalizations:        c.t.SimpleLocalizeAll("list"),
				Description:              "List the commands available in this server and whether they are enabled",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("List the commands available in this server and whether they are enabled"),
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "enable",
				NameLocalizations:        c.t.SimpleLocalizeAll("enable"),
				Description:              "Enable a command in this server",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Enable a command in this server"),
				Options: []disgodiscord.ApplicationCommandOption{
					commandOption,
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "disable",
				NameLocalizations:        c.t.SimpleLocalizeAll("disable"),
				Description:              "Disable a command in this server",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Disable a command in this server"),
				Options: []disgodiscord.ApplicationCommandOption{
					commandOption,
				},
			},
		},
	}
}

func (c *manageCommandImpl) Name() string {
	return ManageCommandName
}

func (c *manageCommandImpl) Scope() CommandScope {
	return CommandScopeGlobal
}

func (c *manageCommandImpl) manageableCommands() []Command {
	commands := lo.Filter(c.registry.ListByScope(CommandScopeGuild), func(cmd Command, _ int) bool {
		return cmd.Name() != ManageCommandName
	})

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})

	return commands
}

func (c *manageCommandImpl) commandChoices() []disgodiscord.ApplicationCommandOptionChoiceString {
	return lo.Map(c.manageableCommands(), func(cmd Command, _ int) disgodiscord.ApplicationCommandOptionChoiceString {
		return disgodiscord.ApplicationCommandOptionChoiceString{
			Name:  cmd.Name(),
			Value: cmd.Name(),
		}
	})
}

func (c *manageCommandImpl) handleList(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	disabled, err := c.settings.ListDisabled(ctx, *e.GuildID())
	if err != nil {
		return fmt.Errorf("failed to list disabled commands: %w", err)
	}
	disabledSet := lo.SliceToMap(disabled, func(name string) (string, struct{}) {
		return name, struct{}{}
	})

	lines := lo.Map(c.manageableCommands(), func(cmd Command, _ int) string {
		status := c.t.SimpleLocalize(e.Locale(), "enabled")
		if _, ok := disabledSet[cmd.Name()]; ok {
			status = c.t.SimpleLocalize(e.Locale(), "disabled")
		}

		return fmt.Sprintf("`/%s` - %s", cmd.Name(), status)
	})

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Embeds: []disgodiscord.Embed{
			{
				Title:       c.t.SimpleLocalize(e.Locale(), "Commands"),
				Description: strings.Join(lines, "\n"),
				Color:       0x00ff00,
			},
		},
	})
}

func (c *manageCommandImpl) handleSetDisabled(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate, disabled bool) error {
	data := e.SlashCommandInteractionData()
	commandName := data.String("command")

	if !lo.ContainsBy(c.manageableCommands(), func(cmd Command) bool {
		return cmd.Name() == commandName
	}) {
		return e.CreateMessage(disgodiscord.MessageCreate{
			Flags: disgodiscord.MessageFlagEphemeral,
			Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
				MessageID: "Unknown command: `{{.Command}}`",
				TemplateData: map[string]any{
					"Command": commandName,
				},
			}),
		})
	}

	guildID := *e.GuildID()

	alreadyDisabled, err := c.settings.IsDisabled(ctx, guildID, commandName)
	if err != nil {
		return fmt.Errorf("failed to check command settings: %w", err)
	}
	if alreadyDisabled == disabled {
		messageID := "`/{{.Command}}` is already enabled in this server."
		if disabled {
			messageID = "`/{{.Command}}` is already disabled in this server."
		}

		return e.CreateMessage(disgodiscord.MessageCreate{
			Flags: disgodiscord.MessageFlagEphemeral,
			Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
				MessageID: messageID,
				TemplateData: map[string]any{
					"Command": commandName,
				},
			}),
		})
	}

	if disabled {
		err = c.settings.Disable(ctx, guildID, commandName)
	} else {
		err = c.settings.Enable(ctx, guildID, commandName)
	}
	if err != nil {
		return fmt.Errorf("failed to update command settings: %w", err)
	}

	if err := c.syncer.SyncGuildCommands(ctx, guildID); err != nil {
		return fmt.Errorf("failed to sync commands after updating settings: %w", err)
	}

	messageID := "`/{{.Command}}` is now enabled in this server."
	if disabled {
		messageID = "`/{{.Command}}` is now disabled in this server."
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Content: c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: messageID,
			TemplateData: map[string]any{
				"Command": commandName,
			},
		}),
	})
}
