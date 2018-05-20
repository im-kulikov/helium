package logger

import (
	"io"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

var defaultEchoLogger = new(echoLogger)

// Null is /dev/null emulation
var Null = new(EmptyWriter)

// EmptyWriter struct
type EmptyWriter struct{}

// Write /dev/null emulation
func (EmptyWriter) Write(data []byte) (int, error) { return len(data), nil }

// Echo logger
func Echo() echo.Logger {
	return defaultEchoLogger
}

type echoLogger struct{}

func (e *echoLogger) Output() io.Writer {
	return Null
}

func (e *echoLogger) SetOutput(w io.Writer) {}

func (e *echoLogger) Prefix() string {
	return ""
}

func (e *echoLogger) SetPrefix(p string) {}

func (e *echoLogger) Level() log.Lvl {
	return log.Level()

}

func (e *echoLogger) SetLevel(v log.Lvl) {}

func (e *echoLogger) Print(i ...interface{}) {
	G().Info(i...)
}

func (e *echoLogger) Printf(format string, args ...interface{}) {
	G().Infof(format, args...)
}

func (e *echoLogger) Printj(j log.JSON) {
	G().Infow("echo json log", "json_msg", j)
}

func (e *echoLogger) Debug(i ...interface{}) {
	G().Debug(i...)
}

func (e *echoLogger) Debugf(format string, args ...interface{}) {
	G().Debugf(format, args...)
}

func (e *echoLogger) Debugj(j log.JSON) {
	G().Debugw("echo json log", "json_msg", j)
}

func (e *echoLogger) Info(i ...interface{}) {
	G().Info(i...)
}

func (e *echoLogger) Infof(format string, args ...interface{}) {
	G().Infof(format, args...)
}

func (e *echoLogger) Infoj(j log.JSON) {
	G().Infow("echo json log", "json_msg", j)
}

func (e *echoLogger) Warn(i ...interface{}) {
	G().Warn(i...)
}

func (e *echoLogger) Warnf(format string, args ...interface{}) {
	G().Warnf(format, args...)
}

func (e *echoLogger) Warnj(j log.JSON) {
	G().Warnw("echo json log", "json_msg", j)
}

func (e *echoLogger) Error(i ...interface{}) {
	G().Error(i...)
}

func (e *echoLogger) Errorf(format string, args ...interface{}) {
	G().Errorf(format, args...)
}

func (e *echoLogger) Errorj(j log.JSON) {
	G().Errorw("echo json log", "json_msg", j)
}

func (e *echoLogger) Fatal(i ...interface{}) {
	G().Fatal(i...)
}

func (e *echoLogger) Fatalf(format string, args ...interface{}) {
	G().Fatalf(format, args...)
}

func (e *echoLogger) Fatalj(j log.JSON) {
	G().Fatalw("echo json log", "json_msg", j)
}

func (e *echoLogger) Panic(i ...interface{}) {
	G().Panic(i...)
}

func (e *echoLogger) Panicf(format string, args ...interface{}) {
	G().Panicf(format, args...)
}

func (e *echoLogger) Panicj(j log.JSON) {
	G().Panicw("echo json log", "json_msg", j)
}
