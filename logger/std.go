package logger

import "go.uber.org/zap"

type (
	// StdLogger interface
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

// Print uses fmt.Sprint to construct and log a message.
func (s stdLogger) Print(v ...interface{}) {
	s.Info(v...)
}

// Printf uses fmt.Sprintf to log a templated message.
func (s stdLogger) Printf(format string, v ...interface{}) {
	s.Infof(format, v...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (s stdLogger) Fatal(v ...interface{}) {
	s.SugaredLogger.Fatal(v...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (s stdLogger) Fatalf(format string, v ...interface{}) {
	s.SugaredLogger.Fatalf(format, v...)
}

// NewStdLogger implementation of StdLogger interface
func NewStdLogger(z *zap.Logger) StdLogger {
	return stdLogger{
		SugaredLogger: z.
			WithOptions(zap.AddCallerSkip(1)).
			Sugar(),
	}
}
