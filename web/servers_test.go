package web

import (
	"net/http"
	"testing"

	"github.com/chapsuk/mserv"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func testHTTPHandler() http.Handler {
	return http.NewServeMux()
}

func TestServers(t *testing.T) {
	var (
		z  = zap.L()
		di = dig.New()
		v  = viper.New()
		l  = logger.NewStdLogger(z)
	)

	t.Run("check pprof server", func(t *testing.T) {
		t.Run("without config", func(t *testing.T) {
			params := profileParams{
				Viper:   v,
				Logger:  l,
				Handler: newProfileHandler().Handler,
			}
			serve := newProfileServer(params)
			assert.Nil(t, serve.Server)
		})

		t.Run("with config", func(t *testing.T) {
			v.SetDefault("pprof.address", ":6090")
			params := profileParams{
				Viper:   v,
				Logger:  l,
				Handler: newProfileHandler().Handler,
			}
			serve := newProfileServer(params)
			assert.NotNil(t, serve.Server)
		})
	})

	t.Run("check metrics server", func(t *testing.T) {
		t.Run("without config", func(t *testing.T) {
			params := metricParams{
				Viper:   v,
				Logger:  l,
				Handler: newMetricHandler().Handler,
			}
			serve := newMetricServer(params)
			assert.Nil(t, serve.Server)
		})

		t.Run("with config", func(t *testing.T) {
			v.SetDefault("metrics.address", ":8090")
			params := metricParams{
				Viper:   v,
				Logger:  l,
				Handler: newMetricHandler().Handler,
			}
			serve := newMetricServer(params)
			assert.NotNil(t, serve.Server)
		})
	})

	t.Run("check api server", func(t *testing.T) {
		t.Run("without config", func(t *testing.T) {
			serve := NewAPIServer(v, l, nil)
			assert.Nil(t, serve.Server)
		})

		t.Run("without handler", func(t *testing.T) {
			v.SetDefault("api.address", ":8090")
			serve := NewAPIServer(v, l, nil)
			assert.Nil(t, serve.Server)
		})

		t.Run("should be ok", func(t *testing.T) {
			v.SetDefault("api.address", ":8090")
			serve := NewAPIServer(v, l, testHTTPHandler())
			assert.NotNil(t, serve.Server)
		})
	})

	t.Run("check multi server", func(t *testing.T) {
		v.SetDefault("pprof.address", ":6090")
		v.SetDefault("metrics.address", ":8090")
		v.SetDefault("api.address", ":8090")

		mod := module.Module{
			{Constructor: func() *viper.Viper { return v }},
			{Constructor: func() logger.StdLogger { return l }},
			{Constructor: func() http.Handler { return testHTTPHandler() }},
		}.Append(
			ServersModule,
			ProfileHandlerModule,
			MetricHandlerModule,
		)

		err := module.Provide(di, mod)
		assert.NoError(t, err)
		err = di.Invoke(func(serve mserv.Server) {
			assert.IsType(t, &mserv.MultiServer{}, serve)
		})
		assert.NoError(t, err)
	})
}
