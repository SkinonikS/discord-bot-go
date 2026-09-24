package cli

import (
	infraservice "github.com/SkinonikS/discord-bot-go/internal/infra"
	"github.com/SkinonikS/discord-bot-go/internal/infra/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/service"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewApplication(params foundation.ModuleParams) (*fx.App, *cli.Command) {
	var cmd *cli.Command
	app := fx.New(
		infraservice.NewModule(params),
		sharedservice.NewModule(),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))}
		}),
		fx.Decorate(func(*zap.Logger) *zap.Logger {
			return zap.NewNop()
		}),
		fx.Populate(&cmd),
	)
	return app, cmd
}
