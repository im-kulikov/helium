package logger

import "go.uber.org/zap"

type (
	StdLogger interface {
		Fatal(v ...interface{})
		Fatalf(format string, v ...interface{})
		Print(v ...interface{})
		Printf(format string, v ...interface{})
	}

	stdLogger struct {
		*zap.SugaredLogger
	}
)

func (s stdLogger) Print(v ...interface{}) {
	s.Info(v...)
}

func (s stdLogger) Printf(format string, v ...interface{}) {
	s.Infof(format, v...)
}

func (s stdLogger) Fatal(v ...interface{}) {
	s.SugaredLogger.Fatal(v...)
}

func (s stdLogger) Fatalf(format string, v ...interface{}) {
	s.SugaredLogger.Fatalf(format, v...)
}

func NewStdLogger(z *zap.Logger) StdLogger {
	l := zap.New(
		z.Core(),
		zap.AddCallerSkip(1),
	)

	return stdLogger{
		SugaredLogger: l.Sugar(),
	}
}
