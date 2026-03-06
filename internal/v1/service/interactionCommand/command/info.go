package command

import (
	"context"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"go.uber.org/fx"
)

type Info struct {
	buildInfo *foundation.BuildInfo
	upTime    discord.UpTime
	config    *config.Config
}

type InfoParams struct {
	fx.In
	Config    *config.Config
	BuildInfo *foundation.BuildInfo
	UpTime    discord.UpTime
}

func NewInfo(p InfoParams) *Info {
	return &Info{
		config:    p.Config,
		upTime:    p.UpTime,
		buildInfo: p.BuildInfo,
	}
}

func (c *Info) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	uptime := time.Since(c.upTime.Time()).Round(time.Second)

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags: disgodiscord.MessageFlagEphemeral,
		Embeds: []disgodiscord.Embed{
			{
				Title:       "Bot Information",
				URL:         c.config.Repository,
				Description: "Details about this bot",
				Color:       0x00ff00,
				Fields: []disgodiscord.EmbedField{
					{Name: "Tag", Value: c.buildInfo.Tag(), Inline: new(true)},
					{Name: "Build time", Value: c.buildInfo.BuildTime(), Inline: new(true)},
					{Name: "Commit", Value: c.buildInfo.Commit(), Inline: new(true)},
					{Name: "UpTime", Value: uptime.String(), Inline: new(true)},
					{Name: "Repository", Value: c.config.Repository, Inline: new(true)},
				},
			},
		},
	}, disgorest.WithCtx(ctx))
}

func (c *Info) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:        c.Name(),
		Description: "Show info about the bot.",
	}
}

func (c *Info) Name() string {
	return "info"
}
