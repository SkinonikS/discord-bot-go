package command

import (
	"context"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
)

type Info struct {
	buildInfo *foundation.BuildInfo
	upTime    time.Time
	config    *config.Config
}

type InfoParams struct {
	fx.In
	Config    *config.Config
	BuildInfo *foundation.BuildInfo
	UpTime    time.Time `name:"discord_uptime"`
}

func NewInfo(p InfoParams) *Info {
	return &Info{
		config:    p.Config,
		upTime:    p.UpTime,
		buildInfo: p.BuildInfo,
	}
}

func (c *Info) Execute(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate) error {
	uptime := time.Since(c.upTime).Round(time.Second)

	embed := &discordgo.MessageEmbed{
		Title:       "Bot Information",
		URL:         c.config.Repository,
		Description: "Details about this bot",
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Tag", Value: c.buildInfo.Tag(), Inline: true},
			{Name: "Build time", Value: c.buildInfo.BuildTime(), Inline: true},
			{Name: "Commit", Value: c.buildInfo.Commit(), Inline: true},
			{Name: "UpTime", Value: uptime.String(), Inline: true},
			{Name: "Repository", Value: c.config.Repository, Inline: true},
			{Name: "Hash", Value: c.buildInfo.Hash()},
		},
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx))
}

func (c *Info) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Show info about the bot.",
	}
}

func (c *Info) ForOwnerOnly() bool {
	return false
}

func (c *Info) Name() string {
	return "info"
}
