package discord

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/readiness"
	disgobot "github.com/disgoorg/disgo/bot"
	disgogateway "github.com/disgoorg/disgo/gateway"
	"go.uber.org/fx"
)

type readinessHandlerImpl struct {
	discordClient *disgobot.Client
}

type ReadinessHandlerParams struct {
	fx.In

	DiscordClient *disgobot.Client
}

func NewReadinessHandler(p ReadinessHandlerParams) readiness.Handler {
	return &readinessHandlerImpl{
		discordClient: p.DiscordClient,
	}
}

func (r *readinessHandlerImpl) IsReady() bool {
	if r.discordClient.HasGateway() {
		return r.discordClient.Gateway.Status() == disgogateway.StatusReady
	}
	return false
}

func (r *readinessHandlerImpl) IsHealthy() bool {
	return r.IsReady()
}
