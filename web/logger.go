package web

import (
	"io"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

type (
	echoLogger struct {
		l *zap.SugaredLogger
	}

	// EmptyWriter struct
	EmptyWriter struct{}
)

// Null is /dev/null emulation
var Null = new(EmptyWriter)

// Write /dev/null emulation
func (EmptyWriter) Write(data []byte) (int, error) { return len(data), nil }

// NewLogger converts zap.Logger to echo.Logger
func NewLogger(log *zap.Logger) echo.Logger {
	return &echoLogger{
		l: log.
			WithOptions(zap.AddCallerSkip(1)).
			Sugar(),
	}
}

// Output writer
func (e *echoLogger) Output() io.Writer {
	return Null
}

// SetOutput do nothing
func (e *echoLogger) SetOutput(w io.Writer) {}

// Prefix do nothing
func (e *echoLogger) Prefix() string {
	return ""
}

// SetPrefix do nothing
func (e *echoLogger) SetPrefix(p string) {}

// Level do nothing
func (e *echoLogger) Level() log.Lvl {
	return log.DEBUG
}

// SetLevel do nothing
func (e *echoLogger) SetLevel(v log.Lvl) {}

// Print uses fmt.Sprint to construct and log a message.
func (e *echoLogger) Print(i ...interface{}) {
	e.l.Info(i...)
}

// Printf uses fmt.Sprintf to log a templated message.
func (e *echoLogger) Printf(format string, args ...interface{}) {
	e.l.Infof(format, args...)
}

// Printj logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (e *echoLogger) Printj(j log.JSON) {
	e.l.Infow("echo json log", "json_msg", j)
}

// Debug uses fmt.Sprint to construct and log a message.
func (e *echoLogger) Debug(i ...interface{}) {
	e.l.Debug(i...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (e *echoLogger) Debugf(format string, args ...interface{}) {
	e.l.Debugf(format, args...)
}

// Debugj logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func (e *echoLogger) Debugj(j log.JSON) {
	e.l.Debugw("echo json log", "json_msg", j)
}

// Info uses fmt.Sprint to construct and log a message.
func (e *echoLogger) Info(i ...interface{}) {
	e.l.Info(i...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (e *echoLogger) Infof(format string, args ...interface{}) {
	e.l.Infof(format, args...)
}

// Infoj logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (e *echoLogger) Infoj(j log.JSON) {
	e.l.Infow("echo json log", "json_msg", j)
}

// Warn uses fmt.Sprint to construct and log a message.
func (e *echoLogger) Warn(i ...interface{}) {
	e.l.Warn(i...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (e *echoLogger) Warnf(format string, args ...interface{}) {
	e.l.Warnf(format, args...)
}

// Warnj logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (e *echoLogger) Warnj(j log.JSON) {
	e.l.Warnw("echo json log", "json_msg", j)
}

// Error uses fmt.Sprint to construct and log a message.
func (e *echoLogger) Error(i ...interface{}) {
	e.l.Error(i...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (e *echoLogger) Errorf(format string, args ...interface{}) {
	e.l.Errorf(format, args...)
}

// Errorj logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (e *echoLogger) Errorj(j log.JSON) {
	e.l.Errorw("echo json log", "json_msg", j)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (e *echoLogger) Fatal(i ...interface{}) {
	e.l.Fatal(i...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (e *echoLogger) Fatalf(format string, args ...interface{}) {
	e.l.Fatalf(format, args...)
}

// Fatalj logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (e *echoLogger) Fatalj(j log.JSON) {
	e.l.Fatalw("echo json log", "json_msg", j)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (e *echoLogger) Panic(i ...interface{}) {
	e.l.Panic(i...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (e *echoLogger) Panicf(format string, args ...interface{}) {
	e.l.Panicf(format, args...)
}

// Panicj logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (e *echoLogger) Panicj(j log.JSON) {
	e.l.Panicw("echo json log", "json_msg", j)
}
