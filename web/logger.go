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

func NewLogger(log *zap.Logger) echo.Logger {
	l := zap.New(
		log.Core(),
		zap.AddCallerSkip(1),
	)

	return &echoLogger{
		l: l.Sugar(),
	}
}

func (e *echoLogger) Output() io.Writer {
	return Null
}

func (e *echoLogger) SetOutput(w io.Writer) {}

func (e *echoLogger) Prefix() string {
	return ""
}

func (e *echoLogger) SetPrefix(p string) {}

func (e *echoLogger) Level() log.Lvl {
	return log.DEBUG
}

func (e *echoLogger) SetLevel(v log.Lvl) {}

func (e *echoLogger) Print(i ...interface{}) {
	e.l.Info(i...)
}

func (e *echoLogger) Printf(format string, args ...interface{}) {
	e.l.Infof(format, args...)
}

func (e *echoLogger) Printj(j log.JSON) {
	e.l.Infow("echo json log", "json_msg", j)
}

func (e *echoLogger) Debug(i ...interface{}) {
	e.l.Debug(i...)
}

func (e *echoLogger) Debugf(format string, args ...interface{}) {
	e.l.Debugf(format, args...)
}

func (e *echoLogger) Debugj(j log.JSON) {
	e.l.Debugw("echo json log", "json_msg", j)
}

func (e *echoLogger) Info(i ...interface{}) {
	e.l.Info(i...)
}

func (e *echoLogger) Infof(format string, args ...interface{}) {
	e.l.Infof(format, args...)
}

func (e *echoLogger) Infoj(j log.JSON) {
	e.l.Infow("echo json log", "json_msg", j)
}

func (e *echoLogger) Warn(i ...interface{}) {
	e.l.Warn(i...)
}

func (e *echoLogger) Warnf(format string, args ...interface{}) {
	e.l.Warnf(format, args...)
}

func (e *echoLogger) Warnj(j log.JSON) {
	e.l.Warnw("echo json log", "json_msg", j)
}

func (e *echoLogger) Error(i ...interface{}) {
	e.l.Error(i...)
}

func (e *echoLogger) Errorf(format string, args ...interface{}) {
	e.l.Errorf(format, args...)
}

func (e *echoLogger) Errorj(j log.JSON) {
	e.l.Errorw("echo json log", "json_msg", j)
}

func (e *echoLogger) Fatal(i ...interface{}) {
	e.l.Fatal(i...)
}

func (e *echoLogger) Fatalf(format string, args ...interface{}) {
	e.l.Fatalf(format, args...)
}

func (e *echoLogger) Fatalj(j log.JSON) {
	e.l.Fatalw("echo json log", "json_msg", j)
}

func (e *echoLogger) Panic(i ...interface{}) {
	e.l.Panic(i...)
}

func (e *echoLogger) Panicf(format string, args ...interface{}) {
	e.l.Panicf(format, args...)
}

func (e *echoLogger) Panicj(j log.JSON) {
	e.l.Panicw("echo json log", "json_msg", j)
}
