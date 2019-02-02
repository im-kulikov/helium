package nats

import (
	"testing"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/nats-streaming-server/server"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func RunServer(ID string) *server.StanServer {
	s, err := server.RunServer(ID)
	if err != nil {
		panic(err)
	}
	return s
}

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

		Convey("should fail for empty config on nats-stremer", func() {
			c, err := NewStreamer(nil)
			So(c, ShouldBeNil)
			So(err, ShouldBeError, ErrEmptyStreamerConfig)
		})

		Convey("should fail client", func() {

			v.SetDefault("nats.url", nats.DefaultURL)

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.Url, ShouldEqual, nats.DefaultURL)

			cli, err := NewConnection(c)
			So(cli, ShouldBeNil)
			So(err, ShouldBeError)
		})

		Convey("should not fail with test server", func() {
			serve := RunServer(nats.DefaultURL)
			defer serve.Shutdown()

			v.SetDefault("nats.url", nats.DefaultURL)

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)

			cli, err := NewConnection(c)
			So(err, ShouldBeNil)
			So(cli, ShouldNotBeNil)
		})

		Convey("should fail with empty config", func() {
			cfg, err := NewDefaultStreamerConfig(v, nil)
			So(cfg, ShouldBeNil)
			So(err, ShouldBeError, ErrEmptyConfig)
		})

		Convey("should fail with empty clusterID", func() {
			v.SetDefault("nats.cluster_id", "")
			cfg, err := NewDefaultStreamerConfig(v, nil)
			So(cfg, ShouldBeNil)
			So(err, ShouldBeError, ErrClusterIDEmpty)
		})

		Convey("should fail with empty clientID", func() {
			v.SetDefault("nats.cluster_id", "myCluster")
			cfg, err := NewDefaultStreamerConfig(v, nil)
			So(cfg, ShouldBeNil)
			So(err, ShouldBeError, ErrClientIDEmpty)
		})

		Convey("should fail on connection empty", func() {
			v.SetDefault("nats.url", nats.DefaultURL)
			v.SetDefault("nats.client_id", "myClient")
			v.SetDefault("nats.cluster_id", "myCluster")

			cfg, err := NewDefaultStreamerConfig(v, nil)
			So(err, ShouldBeNil)

			cfg.Options = nil

			stan, err := NewStreamer(cfg)
			So(stan, ShouldBeNil)
			So(err, ShouldBeError, ErrEmptyConnection)
		})

		Convey("should run streamer client", func() {
			v.SetDefault("nats.client_id", "myClient")
			v.SetDefault("nats.cluster_id", "myCluster")

			// Run a NATS Streaming server
			s := RunServer("myCluster")
			defer s.Shutdown()

			con, err := nats.Connect(nats.DefaultURL)
			So(err, ShouldBeNil)

			defer con.Close()

			cfg, err := NewDefaultStreamerConfig(v, con)
			So(err, ShouldBeNil)

			st, err := NewStreamer(cfg)
			So(err, ShouldBeNil)
			So(st.Close(), ShouldBeNil)
		})
	})
}
