package discord

import (
	"context"
	"log/slog"

	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"github.com/disgoorg/disgo"
	disgobot "github.com/disgoorg/disgo/bot"
	disgocache "github.com/disgoorg/disgo/cache"
	disgogateway "github.com/disgoorg/disgo/gateway"
	disgovoice "github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/godave/golibdave"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In
	Lc        fx.Lifecycle
	Config    *Config
	AppConfig *config.Config
	Log       *zap.Logger
}

func New(p Params) (*disgobot.Client, error) {
	intents, err := p.Config.ParseIntents()
	if err != nil {
		return nil, err
	}

	slogger := slog.New(slogzap.Option{Logger: p.Log, Level: slog.LevelInfo}.NewZapHandler())
	client, err := disgo.New(p.Config.Token,
		disgobot.WithCacheConfigOpts(
			disgocache.WithCaches(disgocache.FlagVoiceStates),
		),
		disgobot.WithGatewayConfigOpts(
			disgogateway.WithIntents(intents),
			disgogateway.WithPresenceOpts(
				disgogateway.WithListeningActivity("I am always watching you"),
			),
		),
		disgobot.WithVoiceManagerConfigOpts(
			disgovoice.WithDaveSessionCreateFunc(golibdave.NewSession),
			disgovoice.WithDaveSessionLogger(slogger),
		),
		disgobot.WithLogger(slogger),
	)
	if err != nil {
		return nil, err
	}

	p.Lc.Append(fx.StartStopHook(
		func(ctx context.Context) error {
			go func() {
				if err := client.OpenGateway(ctx); err != nil {
					p.Log.Panic("can't establish connection with discord gateway", zap.Error(err))
				}
			}()
			p.Log.Debug("connected to discord gateway")
			return nil
		},
		func(ctx context.Context) error {
			client.Close(ctx)
			p.Log.Info("discord gateway connection closed")
			return nil
		},
	))

	return client, nil
}
