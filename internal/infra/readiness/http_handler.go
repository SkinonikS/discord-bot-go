package readiness

import (
	"net/http"

	httpserver "github.com/SkinonikS/discord-bot-go/internal/infra/http_server"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

type httpHandlerImpl struct {
	registry      Registry
	discordClient *disgobot.Client
}

type HTTPHandlerParams struct {
	fx.In

	Registry      Registry
	DiscordClient *disgobot.Client
}

func NewHTTPHandler(p HTTPHandlerParams) httpserver.Handler {
	return &httpHandlerImpl{
		registry:      p.Registry,
		discordClient: p.DiscordClient,
	}
}

func (h *httpHandlerImpl) Register(app *fiber.App) error {
	app.Get("/livez", func(c fiber.Ctx) error {
		statusCode := http.StatusOK
		isReady := h.registry.IsHealthy()
		if !isReady {
			statusCode = http.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(map[string]any{
			"isHealthy": isReady,
		})
	})

	app.Get("/readyz", func(c fiber.Ctx) error {
		statusCode := http.StatusOK
		isReady := h.registry.IsReady()
		if !isReady {
			statusCode = http.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(map[string]any{
			"isReady": isReady,
		})
	})

	return nil
}
