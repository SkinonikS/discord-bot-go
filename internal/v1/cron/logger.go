package cron

import (
	"strings"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type loggerImpl struct {
	log *zap.SugaredLogger
}

func NewLogger(log *zap.SugaredLogger) gocron.Logger {
	return &loggerImpl{
		log: log,
	}
}

func (l *loggerImpl) Debug(msg string, args ...any) {
	l.log.Debugf(l.removePrefix(msg), args...)
}

func (l *loggerImpl) Error(msg string, args ...any) {
	l.log.Errorf(l.removePrefix(msg), args...)
}

func (l *loggerImpl) Info(msg string, args ...any) {
	l.log.Infof(l.removePrefix(msg), args...)
}

func (l *loggerImpl) Warn(msg string, args ...any) {
	l.log.Warnf(l.removePrefix(msg), args...)
}

func (l *loggerImpl) removePrefix(msg string) string {
	return strings.TrimPrefix(msg, "gocron: ")
}
