package discord

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/pkg/v1/discord"
	disgobot "github.com/disgoorg/disgo/bot"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type WorkerPoolParams struct {
	fx.In
	Client         *disgobot.Client
	Lc             fx.Lifecycle
	Log            *zap.Logger
	Config         *Config
	EventListeners []disgobot.EventListener `group:"discord_event_listeners"`
}

func NewWorkerPool(p WorkerPoolParams) *discord.WorkerPool {
	wp := discord.NewWorkerPool(discord.WorkerPoolParams{
		WorkerCount:    p.Config.WorkerCount,
		EventListeners: p.EventListeners,
		Log:            p.Log.Sugar(),
	})

	p.Client.AddEventListeners(wp)
	p.Lc.Append(fx.StartStopHook(
		func(context.Context) error {
			if err := wp.Start(); err != nil {
				return err
			}
			p.Log.Info("worker pool started")
			return nil
		},
		func(context.Context) error {
			if err := wp.Stop(); err != nil {
				return err
			}
			p.Log.Info("worker pool stopped")
			return nil
		},
	))

	return wp
}
