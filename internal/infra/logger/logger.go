package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Params struct {
	fx.In

	Lc     fx.Lifecycle
	Config *Config
}

func New(p Params) (*zap.Logger, error) {
	if p.Config.Disable {
		return zap.NewNop(), nil
	}

	switch p.Config.Format {
	case FormatPretty:
		return prettyLogger(p.Config, p.Lc)
	case FormatJSON:
		return jsonLogger(p.Config, p.Lc)
	default:
		return nil, fmt.Errorf("unknown format: %s", p.Config.Format)
	}
}

func jsonLogger(cfg *Config, lc fx.Lifecycle) (*zap.Logger, error) {
	zapCfg := zap.NewProductionConfig()

	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	zapCfg.Level = zap.NewAtomicLevelAt(cfg.Level.AsZap())

	logger, err := zapCfg.Build(
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCaller(),
	)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.StopHook(func(_ context.Context) error {
		_ = logger.Sync()
		return nil
	}))

	return logger, nil
}

func prettyLogger(cfg *Config, lc fx.Lifecycle) (*zap.Logger, error) {
	encoderCfg := prettyconsole.NewEncoderConfig()
	encoderCfg.EncodeTime = prettyconsole.DefaultTimeEncoder(time.DateTime)

	core := zapcore.NewCore(
		prettyconsole.NewEncoder(encoderCfg),
		os.Stdout,
		cfg.Level.AsZap(),
	)

	logger := zap.New(core,
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCaller(),
	)

	lc.Append(fx.StopHook(func(_ context.Context) error {
		_ = logger.Sync()
		return nil
	}))

	return logger, nil
}
