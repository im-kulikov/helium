package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

// Config for logger
type Config struct {
	AppName    string
	AppVersion string
	Level      string
	Format     string
}

// Panic when logger not inited
func Panic(err error) {
	fmt.Printf(`{"app_name": "%s", "app_version": "%s", "msg": "start app error", "error": "%s"}`, "", "", err.Error())
	os.Exit(1)
}

// Init logger
func Init(lcfg *Config, customInit func(lcfg *Config) error) error {
	if customInit != nil {
		return customInit(lcfg)
	}

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stdout"}

	cfg.Encoding = lcfg.Format
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var lvl zapcore.Level
	lvl.Set(lcfg.Level)
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	l, err := cfg.Build()
	if err != nil {
		return err
	}

	if lcfg.Format == "console" {
		logger = l.Sugar()
		return nil
	}

	logger = l.Sugar().With(
		"app_name", lcfg.AppName,
		"app_version", lcfg.AppVersion,
	)

	return nil
}

// G global logger
func G() *zap.SugaredLogger {
	if logger == nil {
		return zap.S()
	}
	return logger
}
