package logger

import (
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/im-kulikov/helium/settings"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestZapLogger(t *testing.T) {
	Convey("ZapLogger test suite", t, func(c C) {
		v := viper.New()

		c.Convey("check logger config", func(c C) {
			c.Convey("empty config", func(c C) {
				cfg := NewLoggerConfig(v)
				c.So(cfg.Level, ShouldBeZeroValue)
				c.So(cfg.Format, ShouldBeZeroValue)
			})

			c.Convey("setup config", func(c C) {
				v.SetDefault("logger.level", "info")
				v.SetDefault("logger.format", "console")
				cfg := NewLoggerConfig(v)
				c.So(cfg.Level, ShouldEqual, "info")
				c.So(cfg.Format, ShouldEqual, "console")
			})

			c.Convey("config safely", func(c C) {
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

					c.So(cfg.SafeLevel(), ShouldEqual, item)
					c.So(cfg.SafeFormat(), ShouldEqual, "json")
				}

				for _, item := range formats {
					v.SetDefault("logger.level", "bad")
					v.SetDefault("logger.format", item)
					cfg := NewLoggerConfig(v)
					if item == "bad" {
						item = "json"
					}

					c.So(cfg.SafeLevel(), ShouldEqual, "info")
					c.So(cfg.SafeFormat(), ShouldEqual, item)
				}
			})
		})

		c.Convey("check logger", func(c C) {
			c.Convey("all ok", func(c C) {
				cfg := NewLoggerConfig(v)
				log, err := NewLogger(cfg, &settings.Core{})
				c.So(err, ShouldBeNil)
				c.So(log, ShouldNotBeNil)
			})

			c.Convey("should fail on level", func(c C) {
				v.SetDefault("logger.level", "bad")
				cfg := NewLoggerConfig(v)
				log, err := NewLogger(cfg, &settings.Core{})
				c.So(err, ShouldBeError)
				c.So(log, ShouldBeNil)
			})

			c.Convey("should fail on stdout", func(c C) {
				monkey.Patch(zap.Open, func(paths ...string) (zapcore.WriteSyncer, func(), error) {
					return nil, nil, errors.New("test")
				})

				defer monkey.Unpatch(zap.Open)

				v.SetDefault("logger.level", "info")
				cfg := NewLoggerConfig(v)
				log, err := NewLogger(cfg, &settings.Core{})
				c.So(err, ShouldBeError)
				c.So(log, ShouldBeNil)
			})

			c.Convey("check sugared", func(c C) {
				v.SetDefault("logger.level", "info")
				cfg := NewLoggerConfig(v)
				log, err := NewLogger(cfg, &settings.Core{})
				c.So(err, ShouldBeNil)
				c.So(log, ShouldNotBeNil)
				sug := NewSugaredLogger(log)
				c.So(sug, ShouldNotBeNil)
			})
		})
	})
}
