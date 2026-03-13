package interactionCommand

import (
	"context"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/v1/translator"
	"github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"go.uber.org/fx"
)

const (
	InfoCommandName = "info"
)

type infoCommandImpl struct {
	t         translator.Translator
	buildInfo *foundation.BuildInfo
	upTime    discord.UpTime
	config    *config.Config
}

type InfoCommandParams struct {
	fx.In

	T         translator.Translator
	Config    *config.Config
	BuildInfo *foundation.BuildInfo
	UpTime    discord.UpTime
}

func NewInfoCommand(p InfoCommandParams) Command {
	return &infoCommandImpl{
		t:         p.T,
		config:    p.Config,
		upTime:    p.UpTime,
		buildInfo: p.BuildInfo,
	}
}

func (c *infoCommandImpl) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
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
					{Name: c.t.SimpleLocalize(e.Locale(), "Tag"), Value: c.buildInfo.Tag(), Inline: new(true)},
					{Name: c.t.SimpleLocalize(e.Locale(), "Build time"), Value: c.buildInfo.BuildTime(), Inline: new(true)},
					{Name: c.t.SimpleLocalize(e.Locale(), "Commit"), Value: c.buildInfo.Commit(), Inline: new(true)},
					{Name: c.t.SimpleLocalize(e.Locale(), "UpTime"), Value: uptime.String(), Inline: new(true)},
					{Name: c.t.SimpleLocalize(e.Locale(), "Repository"), Value: c.config.Repository, Inline: new(true)},
				},
			},
		},
	}, disgorest.WithCtx(ctx))
}

func (c *infoCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		NameLocalizations:        c.t.SimpleLocalizeAll(c.Name()),
		Description:              "Show info about the bot.",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("Show info about the bot."),
	}
}

func (c *infoCommandImpl) Name() string {
	return InfoCommandName
}
