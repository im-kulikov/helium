package nats

import (
	"fmt"
	"runtime"
	"testing"

	gnatsd "github.com/nats-io/gnatsd/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func TestNewDefaultConfig(t *testing.T) {
	Convey("Check nats module", t, func() {
		v := viper.New()

		Convey("must fail on empty", func() {
			c, err := NewDefaultConfig(v)
			So(c, ShouldBeNil)
			So(err, ShouldBeError)
		})

		Convey("servers should be nil", func() {
			v.SetDefault("nats.url", "something")

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.Servers, ShouldBeNil)
		})

		Convey("servers should be slice of string", func() {
			v.SetDefault("nats.servers", "something")

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.Servers, ShouldHaveLength, 1)
			So(c.Servers[0], ShouldEqual, "something")
		})

		Convey("should be ok", func() {
			url := "something"
			v.SetDefault("nats.url", url)

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.Url, ShouldEqual, url)
		})

		Convey("should fail for empty config", func() {
			c, err := NewConnection(nil)
			So(c, ShouldBeNil)
			So(err, ShouldBeError, ErrEmptyConfig)
		})

		Convey("should fail client", func() {
			port := 8368
			url := fmt.Sprintf("nats://localhost:%d", port)
			v.SetDefault("nats.url", url)

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.Url, ShouldEqual, url)

			cli, err := NewConnection(c)
			So(cli, ShouldBeNil)
			So(err, ShouldBeError)
		})

		Convey("should not fail with test server", func() {
			port := 8368
			url := fmt.Sprintf("nats://localhost:%d", port)
			v.SetDefault("nats.url", url)

			opts := gnatsd.DefaultTestOptions
			opts.Port = port
			serve := gnatsd.RunServer(&opts)
			defer serve.Shutdown()
			go serve.Start()
			runtime.Gosched()

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.Url, ShouldEqual, url)

			cli, err := NewConnection(c)
			So(err, ShouldBeNil)
			So(cli, ShouldNotBeNil)
		})
	})
}
