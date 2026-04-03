package readiness

import (
	"net/http"

	"github.com/SkinonikS/discord-bot-go/internal/v1/httpServer"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type httpHandler struct {
	registry      Registry
	discordClient *disgobot.Client
}

type HTTPHandlerParams struct {
	fx.In

	Registry      Registry
	DiscordClient *disgobot.Client
}

func NewHTTPHandler(p HTTPHandlerParams) httpServer.Handler {
	return &httpHandler{
		registry:      p.Registry,
		discordClient: p.DiscordClient,
	}
}

func (h *httpHandler) Register(engine *gin.Engine) error {
	engine.GET("/healthz", func(c *gin.Context) {
		statusCode := http.StatusOK
		isReady := h.registry.IsHealthy()
		if !isReady {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"isHealthy": isReady,
		})
	})

	engine.GET("/readyz", func(c *gin.Context) {
		statusCode := http.StatusOK
		isReady := h.registry.IsReady()
		if !isReady {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"isReady": isReady,
		})
	})

	return nil
}
