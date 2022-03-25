package logger

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/im-kulikov/helium/settings"
)

// Config for logger.
type Config struct {
	Level        string
	TraceLevel   string
	Format       string
	Debug        bool
	Color        bool
	NoCaller     bool
	FullCaller   bool
	NoDisclaimer bool
	Sampling     *zap.SamplingConfig
}

const (
	defaultSamplingInitial    = 100
	defaultSamplingThereafter = 100
)

// NewLoggerConfig returns logger config.
func NewLoggerConfig(v *viper.Viper) *Config {
	cfg := &Config{
		Debug:        v.GetBool("debug"),
		Level:        v.GetString("logger.level"),
		TraceLevel:   v.GetString("logger.trace_level"),
		Format:       v.GetString("logger.format"),
		Color:        v.GetBool("logger.color"),
		NoCaller:     v.GetBool("logger.no_caller"),
		FullCaller:   v.GetBool("logger.full_caller"),
		NoDisclaimer: v.GetBool("logger.no_disclaimer"),
	}

	if v.IsSet("logger.sampling") {
		cfg.Sampling = &zap.SamplingConfig{
			Initial:    defaultSamplingInitial,
			Thereafter: defaultSamplingThereafter,
		}

		if val := v.GetInt("logger.sampling.initial"); val > 0 {
			cfg.Sampling.Initial = val
		}

		if val := v.GetInt("logger.sampling.thereafter"); val > 0 {
			cfg.Sampling.Thereafter = val
		}
	}

	return cfg
}

// SafeLevel returns valid logger level or default.
func SafeLevel(lvl string, defaultLvl zapcore.Level) zap.AtomicLevel {
	switch strings.ToLower(lvl) {
	case zapcore.DebugLevel.String():
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case zapcore.InfoLevel.String():
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case zapcore.WarnLevel.String():
		return zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case zapcore.ErrorLevel.String():
		return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case zapcore.PanicLevel.String():
		return zap.NewAtomicLevelAt(zapcore.PanicLevel)
	case zapcore.FatalLevel.String():
		return zap.NewAtomicLevelAt(zapcore.FatalLevel)
	default:
		return zap.NewAtomicLevelAt(defaultLvl)
	}
}

// SafeFormat returns valid logger output format use json by default.
// nolint:goconst
func (c Config) SafeFormat() string {
	switch c.Format {
	case "console", "json":
		return c.Format
	default:
		return "json"
	}
}

// NewSugaredLogger converts from zap.Logger.
func NewSugaredLogger(log *zap.Logger) *zap.SugaredLogger {
	return log.Sugar()
}

// NewLogger init logger.
func NewLogger(lcfg *Config, app *settings.Core) (*zap.Logger, error) {
	var cfg zap.Config
	if lcfg.Debug {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	if lcfg.Sampling != nil {
		cfg.Sampling = lcfg.Sampling
	}

	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stdout"}

	cfg.Encoding = lcfg.SafeFormat()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if lcfg.Color {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if lcfg.NoCaller {
		cfg.DisableCaller = true
	}

	if lcfg.FullCaller {
		cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}

	cfg.Level = SafeLevel(lcfg.Level, zapcore.InfoLevel)
	traceLevel := SafeLevel(lcfg.TraceLevel, zapcore.WarnLevel)

	l, err := cfg.Build(
		// enable trace only for current log-level
		zap.AddStacktrace(traceLevel))
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
