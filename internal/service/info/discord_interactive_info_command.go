package info

import (
	"context"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/infra/config"
	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/infra/translator"
	"github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"go.uber.org/fx"
)

type interactiveInfoCommandImpl struct {
	t         translator.Translator
	buildInfo foundation.BuildInfo
	upTime    discord.UpTime
	config    *config.Config
}

type InteractiveInfoCommandParams struct {
	fx.In

	T         translator.Translator
	Config    *config.Config
	BuildInfo foundation.BuildInfo
	UpTime    discord.UpTime
}

func NewInteractiveInfoCommand(p InteractiveInfoCommandParams) interactioncommand.Command { //nolint:gocritic // fx.In params must be passed by value
	return &interactiveInfoCommandImpl{
		t:         p.T,
		config:    p.Config,
		upTime:    p.UpTime,
		buildInfo: p.BuildInfo,
	}
}

func (c *interactiveInfoCommandImpl) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	uptime := time.Since(c.upTime.Time()).Round(time.Second)

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Embeds: []disgodiscord.Embed{
			{
				Title:       c.t.SimpleLocalize(e.Locale(), "Bot Information"),
				URL:         c.config.Repository,
				Description: c.t.SimpleLocalize(e.Locale(), "Details about this bot"),
				Color:       0x00ff00,
				Fields: []disgodiscord.EmbedField{
					{Name: c.t.SimpleLocalize(e.Locale(), "Tag"), Value: c.buildInfo.Tag, Inline: new(true)},
					{Name: c.t.SimpleLocalize(e.Locale(), "Build time"), Value: c.buildInfo.BuildTime, Inline: new(true)},
					{Name: c.t.SimpleLocalize(e.Locale(), "Commit"), Value: c.buildInfo.Commit, Inline: new(true)},
					{Name: c.t.SimpleLocalize(e.Locale(), "UpTime"), Value: uptime.String(), Inline: new(true)},
					{Name: c.t.SimpleLocalize(e.Locale(), "Repository"), Value: c.config.Repository, Inline: new(true)},
				},
			},
		},
	}, disgorest.WithCtx(ctx))
}

func (c *interactiveInfoCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		NameLocalizations:        c.t.SimpleLocalizeAll(c.Name()),
		Description:              "Show info about the bot.",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("Show info about the bot."),
	}
}

func (c *interactiveInfoCommandImpl) Name() string {
	return "info"
}

func (c *interactiveInfoCommandImpl) Scope() interactioncommand.CommandScope {
	return interactioncommand.CommandScopeGlobal
}
