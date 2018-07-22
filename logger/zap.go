package logger

import (
	"github.com/im-kulikov/helium/settings"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config for logger
type Config struct {
	Level  string
	Format string
}

func NewLoggerConfig(v *viper.Viper) *Config {
	return &Config{
		Level:  v.GetString("log.level"),
		Format: v.GetString("log.format"),
	}
}

// SafeLevel returns valid logger level
// use info level by default
func (c Config) SafeLevel() string {
	switch c.Level {
	case "debug", "DEBUG":
	case "info", "INFO":
	case "warn", "WARN":
	case "error", "ERROR":
	case "panic", "PANIC":
	case "fatal", "FATAL":
	default:
		return "info"
	}
	return c.Level
}

// SafeFormat returns valid logger output format
// use json by default
func (c Config) SafeFormat() string {
	switch c.Format {
	case "console":
	case "json":
	default:
		return "json"
	}
	return c.Format
}

// Init logger
func NewLogger(lcfg *Config, app *settings.App) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stdout"}

	cfg.Encoding = lcfg.SafeFormat()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var lvl zapcore.Level
	if err := lvl.Set(lcfg.Level); err != nil {
		return nil, err
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	if lcfg.SafeFormat() == "console" {
		return l, nil
	}

	return l.With(
		zap.String("app_name", app.Name),
		zap.String("app_version", app.BuildVersion),
	), nil
}
