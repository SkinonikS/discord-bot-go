package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/config"
	"github.com/bwmarrin/discordgo"
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

type Result struct {
	fx.Out
	Session *discordgo.Session
	UpTime  time.Time `name:"discord_uptime"`
}

func New(p Params) (Result, error) {
	discordgo.Logger = func(msgL, caller int, format string, a ...interface{}) {
		p.Log.Sugar().Debugf(format, a...)
	}

	session, err := discordgo.New(fmt.Sprintf("Bot %s", p.Config.Token))
	if err != nil {
		return Result{}, err
	}

	intents, err := p.Config.ParseIntents()
	if err != nil {
		return Result{}, err
	}
	session.Identify.Intents = intents

	session.Identify.Presence.Game = discordgo.Activity{
		Name: "I am always watching you",
	}

	p.Lc.Append(fx.StartStopHook(
		func(ctx context.Context) error {
			p.Log.Debug("connecting to discord gateway")
			go func() {
				if err := session.Open(); err != nil {
					p.Log.Panic("can't establish connection to discord gateway", zap.Error(err))
				} else {
					p.Log.Info("discord gateway connection established",
						zap.String("user", fmt.Sprintf("%s#%s", session.State.User.Username, session.State.User.Discriminator)),
						zap.String("session", session.State.SessionID),
						zap.Int("guilds", len(session.State.Guilds)),
					)
				}
			}()

			return nil
		},
		func(ctx context.Context) error {
			if err := session.Close(); err != nil {
				return err
			}

			p.Log.Info("discord connection closed")
			return nil
		},
	))

	return Result{
		Session: session,
		UpTime:  time.Now(),
	}, nil
}
