package migrator

import "go.uber.org/zap"

type Logger struct {
	log *zap.SugaredLogger
}

func NewLogger(log *zap.Logger) *Logger {
	return &Logger{
		log: log.Sugar(),
	}
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log.Fatalf(format, v...)
}

func (l *Logger) Printf(format string, v ...any) {
	l.log.Debugf(format, v...)
}
