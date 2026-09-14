package migrator

import (
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type loggerImpl struct {
	log *zap.SugaredLogger
}

func NewLogger(log *zap.SugaredLogger) goose.Logger {
	return &loggerImpl{
		log: log,
	}
}

func (l *loggerImpl) Fatalf(format string, v ...any) {
	l.log.Fatalf(format, v...)
}

func (l *loggerImpl) Printf(format string, v ...any) {
	l.log.Debugf(format, v...)
}
