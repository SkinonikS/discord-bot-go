package discord

import (
	"fmt"

	disgoevents "github.com/disgoorg/disgo/events"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type eventListener struct {
	log *zap.SugaredLogger
}

type EventListenerParams struct {
	fx.In

	Log *zap.Logger
}

func NewEventListener(p EventListenerParams) *disgoevents.ListenerAdapter {
	el := &eventListener{
		log: p.Log.Sugar(),
	}

	return &disgoevents.ListenerAdapter{
		OnReady: el.Ready,
	}
}

func (el *eventListener) Ready(e *disgoevents.Ready) {
	el.log.Infow("discord gateway connection established",
		zap.String("user", fmt.Sprintf("%s#%s", e.User.Username, e.User.Discriminator)),
		zap.Int64("intents", int64(e.Client().Gateway.Intents())),
	)
}
