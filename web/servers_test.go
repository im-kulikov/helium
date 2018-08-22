package web

import (
	"net/http"
	"testing"

	"github.com/chapsuk/mserv"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/labstack/echo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func TestServers(t *testing.T) {
	Convey("Servers test suite", t, func() {
		v := viper.New()
		z := zap.L()
		l := logger.NewStdLogger(z)
		di := dig.New()

		Convey("check pprof server", func() {
			Convey("without config", func() {
				serve := NewPprofServer(v, l)
				So(serve.Server, ShouldBeNil)
			})

			Convey("with config", func() {
				v.SetDefault("pprof.address", ":6090")
				serve := NewPprofServer(v, l)
				So(serve.Server, ShouldNotBeNil)
			})
		})

		Convey("check metrics server", func() {
			Convey("without config", func() {
				serve := NewMetricsServer(v, l)
				So(serve.Server, ShouldBeNil)
			})

			Convey("with config", func() {
				v.SetDefault("metrics.address", ":8090")
				serve := NewMetricsServer(v, l)
				So(serve.Server, ShouldNotBeNil)
			})
		})

		Convey("check api server", func() {
			Convey("without config", func() {
				serve := NewAPIServer(v, l, nil)
				So(serve.Server, ShouldBeNil)
			})

			Convey("without handler", func() {
				v.SetDefault("api.address", ":8090")
				serve := NewAPIServer(v, l, nil)
				So(serve.Server, ShouldBeNil)
			})

			Convey("should be ok", func() {
				v.SetDefault("api.address", ":8090")
				serve := NewAPIServer(v, l, echo.New())
				So(serve.Server, ShouldNotBeNil)
			})
		})

		Convey("check multi server", func() {
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
				{Constructor: func() http.Handler { return echo.New() }},
			}

			err := module.Provide(di, mod)
			So(err, ShouldBeNil)
			err = di.Invoke(func(serve mserv.Server) {
				So(serve, ShouldHaveSameTypeAs, &mserv.MultiServer{})
			})
			So(err, ShouldBeNil)
		})
	})
}
