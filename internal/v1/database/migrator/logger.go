package migrator

import "go.uber.org/zap"

type Logger struct {
	log *zap.SugaredLogger
}

func NewLogger(log *zap.SugaredLogger) *Logger {
	return &Logger{
		log: log,
	}
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log.Fatalf(format, v...)
}

func (l *Logger) Printf(format string, v ...any) {
	l.log.Debugf(format, v...)
}
