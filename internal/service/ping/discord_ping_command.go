package ping

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/internal/infra/translator"
	"github.com/SkinonikS/discord-bot-go/internal/service/interaction_command"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/fx"
)

type discordPingCommandImpl struct {
	t translator.Translator
}

type DiscordPingCommandParams struct {
	fx.In

	T translator.Translator
}

func NewDiscordPingCommand(p DiscordPingCommandParams) interactioncommand.Command {
	return &discordPingCommandImpl{
		t: p.T,
	}
}

func (c *discordPingCommandImpl) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	if err := e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: c.t.SimpleLocalize(e.Locale(), "Pinging..."),
	}, disgorest.WithCtx(ctx)); err != nil {
		return err
	}

	msg, err := e.Client().Rest.GetInteractionResponse(e.ApplicationID(), e.Token(), disgorest.WithCtx(ctx))
	if err != nil {
		return err
	}

	latency := msg.ID.Time().Sub(e.ID().Time()).Milliseconds()

	_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
		Content: new(c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "Pong! {{.Latency}}ms",
			TemplateData: map[string]any{
				"Latency": latency,
			},
		})),
	}, disgorest.WithCtx(ctx))
	return err
}

func (c *discordPingCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		NameLocalizations:        c.t.SimpleLocalizeAll(c.Name()),
		Description:              "Pong!",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("Pong!"),
	}
}

func (c *discordPingCommandImpl) Name() string {
	return "ping"
}

func (c *discordPingCommandImpl) Scope() interactioncommand.CommandScope {
	return interactioncommand.CommandScopeGlobal
}
