package logger

import (
	"github.com/im-kulikov/helium/settings"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config for logger
type Config struct {
	Level        string
	Format       string
	Debug        bool
	Color        bool
	FullCaller   bool
	NoDisclaimer bool
}

// NewLoggerConfig returns logger config
func NewLoggerConfig(v *viper.Viper) *Config {
	return &Config{
		Debug:        v.GetBool("debug"),
		Level:        v.GetString("logger.level"),
		Format:       v.GetString("logger.format"),
		Color:        v.GetBool("logger.color"),
		FullCaller:   v.GetBool("logger.full_caller"),
		NoDisclaimer: v.GetBool("logger.no_disclaimer"),
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

// NewSugaredLogger converts from zap.Logger
func NewSugaredLogger(log *zap.Logger) *zap.SugaredLogger {
	return log.Sugar()
}

// NewLogger init logger
func NewLogger(lcfg *Config, app *settings.Core) (*zap.Logger, error) {
	var cfg zap.Config
	if lcfg.Debug {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stdout"}

	cfg.Encoding = lcfg.SafeFormat()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if lcfg.Color {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if lcfg.FullCaller {
		cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}

	var lvl zapcore.Level
	if err := lvl.Set(lcfg.Level); err != nil {
		return nil, err
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	// don't display app name and version
	if lcfg.NoDisclaimer {
		return l, nil
	}

	return l.With(
		zap.String("app_name", app.Name),
		zap.String("app_version", app.BuildVersion),
	), nil
}
