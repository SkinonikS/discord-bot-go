package readiness

import (
	"net/http"

	"github.com/SkinonikS/discord-bot-go/internal/infra/http_server"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/gin-gonic/gin"
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

func NewHTTPHandler(p HTTPHandlerParams) http_server.Handler {
	return &httpHandlerImpl{
		registry:      p.Registry,
		discordClient: p.DiscordClient,
	}
}

func (h *httpHandlerImpl) Register(engine *gin.Engine) error {
	engine.GET("/livez", func(c *gin.Context) {
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
