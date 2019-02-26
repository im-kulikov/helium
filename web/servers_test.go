package web

import (
	"net/http"
	"testing"

	"github.com/chapsuk/mserv"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func testHTTPHandler() http.Handler {
	return http.NewServeMux()
}

func TestServers(t *testing.T) {
	Convey("Servers test suite", t, func(c C) {
		v := viper.New()
		z := zap.L()
		l := logger.NewStdLogger(z)
		di := dig.New()

		c.Convey("check pprof server", func(c C) {
			c.Convey("without config", func(c C) {
				serve := NewPprofServer(v, l)
				c.So(serve.Server, ShouldBeNil)
			})

			c.Convey("with config", func(c C) {
				v.SetDefault("pprof.address", ":6090")
				serve := NewPprofServer(v, l)
				c.So(serve.Server, ShouldNotBeNil)
			})
		})

		c.Convey("check metrics server", func(c C) {
			c.Convey("without config", func(c C) {
				serve := NewMetricsServer(v, l)
				c.So(serve.Server, ShouldBeNil)
			})

			c.Convey("with config", func(c C) {
				v.SetDefault("metrics.address", ":8090")
				serve := NewMetricsServer(v, l)
				c.So(serve.Server, ShouldNotBeNil)
			})
		})

		c.Convey("check api server", func(c C) {
			c.Convey("without config", func(c C) {
				serve := NewAPIServer(v, l, nil)
				c.So(serve.Server, ShouldBeNil)
			})

			c.Convey("without handler", func(c C) {
				v.SetDefault("api.address", ":8090")
				serve := NewAPIServer(v, l, nil)
				c.So(serve.Server, ShouldBeNil)
			})

			c.Convey("should be ok", func(c C) {
				v.SetDefault("api.address", ":8090")
				serve := NewAPIServer(v, l, testHTTPHandler())
				c.So(serve.Server, ShouldNotBeNil)
			})
		})

		c.Convey("check multi server", func(c C) {
			v.SetDefault("pprof.address", ":6090")
			v.SetDefault("metrics.address", ":8090")
			v.SetDefault("api.address", ":8090")

			mod := module.Module{
				{Constructor: NewPprofServer},
				{Constructor: NewMetricsServer},
				{Constructor: NewAPIServer},
				{Constructor: NewMultiServer},
				{Constructor: func() *viper.Viper { return v }},
				{Constructor: func() logger.StdLogger { return l }},
				{Constructor: func() http.Handler { return testHTTPHandler() }},
			}

			err := module.Provide(di, mod)
			c.So(err, ShouldBeNil)
			err = di.Invoke(func(serve mserv.Server) {
				c.So(serve, ShouldHaveSameTypeAs, &mserv.MultiServer{})
			})
			c.So(err, ShouldBeNil)
		})
	})
}
