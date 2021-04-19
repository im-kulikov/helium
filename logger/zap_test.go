package logger

import (
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/im-kulikov/helium/settings"
)

func TestZapLogger(t *testing.T) {
	t.Run("check logger config", func(t *testing.T) {
		v := viper.New()

		t.Run("empty config", func(t *testing.T) {
			cfg := NewLoggerConfig(v)
			require.Empty(t, cfg.Level)
			require.Empty(t, cfg.Format)
			require.Empty(t, cfg.Debug)
			require.Empty(t, cfg.NoDisclaimer)
		})

		t.Run("setup config", func(t *testing.T) {
			v.SetDefault("debug", true)
			v.SetDefault("logger.no_disclaimer", true)
			v.SetDefault("logger.level", "info")
			v.SetDefault("logger.format", "console")
			v.SetDefault("logger.color", true)
			v.SetDefault("logger.full_caller", true)
			v.SetDefault("logger.sampling.initial", 100)
			v.SetDefault("logger.sampling.thereafter", 100)

			cfg := NewLoggerConfig(v)
			require.Equal(t, "info", cfg.Level)
			require.Equal(t, "console", cfg.Format)

			require.NotNil(t, cfg.Sampling)
			require.Equal(t, 100, cfg.Sampling.Initial)
			require.Equal(t, 100, cfg.Sampling.Thereafter)

			_, err := NewLogger(cfg, &settings.Core{})
			require.NoError(t, err)
		})

		t.Run("setup config (no caller)", func(t *testing.T) {
			v.SetDefault("debug", true)
			v.SetDefault("logger.no_disclaimer", true)
			v.SetDefault("logger.level", "info")
			v.SetDefault("logger.format", "console")
			v.SetDefault("logger.color", true)
			v.SetDefault("logger.no_caller", true)
			v.SetDefault("logger.sampling.initial", 100)
			v.SetDefault("logger.sampling.thereafter", 100)

			cfg := NewLoggerConfig(v)
			require.Equal(t, "info", cfg.Level)
			require.Equal(t, "console", cfg.Format)

			require.NotNil(t, cfg.Sampling)
			require.Equal(t, 100, cfg.Sampling.Initial)
			require.Equal(t, 100, cfg.Sampling.Thereafter)

			_, err := NewLogger(cfg, &settings.Core{})
			require.NoError(t, err)
		})

		t.Run("config safely", func(t *testing.T) {
			levels := []string{
				"bad",
				"debug", "DEBUG",
				"info", "INFO",
				"warn", "WARN",
				"error", "ERROR",
				"panic", "PANIC",
				"fatal", "FATAL",
			}

			formats := []string{"bad", "console", "json"}

			for _, item := range levels {
				v.SetDefault("logger.level", item)
				v.SetDefault("logger.format", "bad")
				cfg := NewLoggerConfig(v)
				if item == "bad" {
					item = "info"
				}

				require.Equal(t, strings.ToLower(item), SafeLevel(item, zapcore.InfoLevel).String())
				require.Equal(t, "json", cfg.SafeFormat())
			}

			for _, item := range formats {
				v.SetDefault("logger.level", "bad")
				v.SetDefault("logger.format", item)
				cfg := NewLoggerConfig(v)
				if item == "bad" {
					item = "json"
				}

				require.Equal(t, "info", SafeLevel(item, zapcore.InfoLevel).String())
				require.Equal(t, item, cfg.SafeFormat())
			}
		})
	})

	t.Run("check logger", func(t *testing.T) {
		t.Run("all ok", func(t *testing.T) {
			v := viper.New()
			v.SetDefault("debug", true)
			v.SetDefault("logger.no_disclaimer", true)

			cfg := NewLoggerConfig(v)
			log, err := NewLogger(cfg, &settings.Core{})
			require.NoError(t, err)
			require.NotNil(t, log)
		})

		t.Run("should not fail on bad level", func(t *testing.T) {
			v := viper.New()
			v.SetDefault("logger.level", "bad")

			cfg := NewLoggerConfig(v)
			log, err := NewLogger(cfg, &settings.Core{})
			require.NoError(t, err)
			require.NotNil(t, log)
		})

		t.Run("should fail on stdout", func(t *testing.T) {
			v := viper.New()

			monkey.Patch(zap.Open, func(paths ...string) (zapcore.WriteSyncer, func(), error) {
				return nil, nil, errors.New("test")
			})

			defer monkey.Unpatch(zap.Open)

			v.SetDefault("logger.level", "info")
			cfg := NewLoggerConfig(v)
			log, err := NewLogger(cfg, &settings.Core{})
			require.Error(t, err)
			require.Nil(t, log)
		})

		t.Run("check sugared", func(t *testing.T) {
			v := viper.New()
			v.SetDefault("logger.level", "info")
			cfg := NewLoggerConfig(v)
			log, err := NewLogger(cfg, &settings.Core{})
			require.NoError(t, err)
			require.NotNil(t, log)
			sug := NewSugaredLogger(log)
			require.NotNil(t, sug)
		})
	})
}
