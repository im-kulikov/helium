package redis

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func TestRedis(t *testing.T) {
	Convey("Redis test suite", t, func() {
		v := viper.New()
		v.AutomaticEnv()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.SetDefault("redis.address", "localhost:6379")

		Convey("Config", func() {
			Convey("should return error when config file is empty", func() {
				cfg, err := NewDefaultConfig(viper.New())
				So(cfg, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("should return config when address exists", func() {
				cfg, err := NewDefaultConfig(v)
				So(err, ShouldBeNil)
				So(cfg, ShouldNotBeNil)
			})
		})

		Convey("Connection", func() {
			cfg, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(cfg, ShouldNotBeNil)

			Convey("should create redis client", func() {
				cli, err := NewConnection(cfg)
				So(err, ShouldBeNil)
				So(cli, ShouldNotBeNil)
			})

			Convey("should return error when address incorrect", func() {
				cfg.Addr = "foo"
				cli, err := NewConnection(cfg)
				So(cli, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})
}
