package logger

import (
	"errors"
	"testing"

	"github.com/bouk/monkey"
	"github.com/im-kulikov/helium/settings"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestZapLogger(t *testing.T) {
	Convey("ZapLogger test suite", t, func() {
		v := viper.New()

		Convey("check logger config", func() {
			Convey("empty config", func() {
				cfg := NewLoggerConfig(v)
				So(cfg.Level, ShouldBeZeroValue)
				So(cfg.Format, ShouldBeZeroValue)
			})

			Convey("setup config", func() {
				v.SetDefault("log.level", "info")
				v.SetDefault("log.format", "console")
				cfg := NewLoggerConfig(v)
				So(cfg.Level, ShouldEqual, "info")
				So(cfg.Format, ShouldEqual, "console")
			})

			Convey("config safely", func() {
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
					v.SetDefault("log.level", item)
					v.SetDefault("log.format", "bad")
					cfg := NewLoggerConfig(v)
					if item == "bad" {
						item = "info"
					}

					So(cfg.SafeLevel(), ShouldEqual, item)
					So(cfg.SafeFormat(), ShouldEqual, "json")
				}

				for _, item := range formats {
					v.SetDefault("log.level", "bad")
					v.SetDefault("log.format", item)
					cfg := NewLoggerConfig(v)
					if item == "bad" {
						item = "json"
					}

					So(cfg.SafeLevel(), ShouldEqual, "info")
					So(cfg.SafeFormat(), ShouldEqual, item)
				}
			})
		})

		Convey("check logger", func() {
			Convey("all ok", func() {
				cfg := NewLoggerConfig(v)
				log, err := NewLogger(cfg, &settings.App{})
				So(err, ShouldBeNil)
				So(log, ShouldNotBeNil)
			})

			Convey("should fail on level", func() {
				v.SetDefault("log.level", "bad")
				cfg := NewLoggerConfig(v)
				log, err := NewLogger(cfg, &settings.App{})
				So(err, ShouldBeError)
				So(log, ShouldBeNil)
			})

			Convey("should fail on stdout", func() {
				monkey.Patch(zap.Open, func(paths ...string) (zapcore.WriteSyncer, func(), error) {
					return nil, nil, errors.New("test")
				})

				defer monkey.Unpatch(zap.Open)

				v.SetDefault("log.level", "info")
				cfg := NewLoggerConfig(v)
				log, err := NewLogger(cfg, &settings.App{})
				So(err, ShouldBeError)
				So(log, ShouldBeNil)
			})

			Convey("check sugared", func() {
				v.SetDefault("log.level", "info")
				cfg := NewLoggerConfig(v)
				log, err := NewLogger(cfg, &settings.App{})
				So(err, ShouldBeNil)
				So(log, ShouldNotBeNil)
				sug := NewSugaredLogger(log)
				So(sug, ShouldNotBeNil)
			})
		})
	})
}
