package interactionCommand

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/internal/v1/translator"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/fx"
)

const (
	PingCommandName = "ping"
)

type pingCommandImpl struct {
	t translator.Translator
}

type PingCommandParams struct {
	fx.In

	T translator.Translator
}

func NewPingCommand(p PingCommandParams) Command {
	return &pingCommandImpl{
		t: p.T,
	}
}

func (c *pingCommandImpl) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
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

func (c *pingCommandImpl) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:                     c.Name(),
		NameLocalizations:        c.t.SimpleLocalizeAll(c.Name()),
		Description:              "Pong!",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("Pong!"),
	}
}

func (c *pingCommandImpl) Name() string {
	return PingCommandName
}
