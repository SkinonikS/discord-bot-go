package bot

import (
	infraservice "github.com/SkinonikS/discord-bot-go/internal/infra"
	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/service"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewApplication(params foundation.ModuleParams) *fx.App {
	return fx.New(
		infraservice.NewModule(params),
		sharedservice.NewModule(),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))}
		}),
	)
}
