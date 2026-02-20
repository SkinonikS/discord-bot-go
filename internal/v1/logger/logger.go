package logger

import (
	"os"
	"time"

	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg *Config) *zap.Logger {
	if cfg.Disable {
		return zap.NewNop()
	}

	encoderCfg := prettyconsole.NewEncoderConfig()
	encoderCfg.EncodeTime = prettyconsole.DefaultTimeEncoder(time.DateTime)
	encoder := prettyconsole.NewEncoder(encoderCfg)

	core := zapcore.NewCore(
		encoder,
		os.Stdout,
		cfg.Level.AsZap(),
	)

	return zap.New(
		core,
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCaller(),
	)
}
