package web

import (
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/chapsuk/mserv"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

var expectResult = []byte("OK")

func testHTTPHandler(assert *require.Assertions) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write(expectResult)
		assert.NoError(err)
	})
	return mux
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
				Viper:  v,
				Logger: l,
			}
			serve := newProfileServer(params)
			require.Nil(t, serve.Server)
		})

		t.Run("with config", func(t *testing.T) {
			v.SetDefault("pprof.address", ":6090")
			params := profileParams{
				Viper:  v,
				Logger: l,
			}
			serve := newProfileServer(params)
			require.NotNil(t, serve.Server)
			require.IsType(t, &mserv.HTTPServer{}, serve.Server)
		})
	})

	t.Run("check metrics server", func(t *testing.T) {
		t.Run("without config", func(t *testing.T) {
			params := metricParams{
				Viper:  v,
				Logger: l,
			}
			serve := newMetricServer(params)
			require.Nil(t, serve.Server)
		})

		t.Run("with config", func(t *testing.T) {
			v.SetDefault("metrics.address", ":8090")
			params := metricParams{
				Viper:  v,
				Logger: l,
			}
			serve := newMetricServer(params)
			require.NotNil(t, serve.Server)
			require.IsType(t, &mserv.HTTPServer{}, serve.Server)
		})
	})

	t.Run("disabled http-server", func(t *testing.T) {
		is := require.New(t)

		v.SetDefault("test-api.disabled", true)

		z, err := zap.NewDevelopment()
		is.NoError(err)

		l := logger.NewStdLogger(z)

		testHTTPHandler(is)

		serve := NewHTTPServer(v, "test-api", testHTTPHandler(is), l)
		is.Nil(serve.Server)
	})

	t.Run("check api server", func(t *testing.T) {
		t.Run("without config", func(t *testing.T) {
			serve := NewAPIServer(v, l, nil)
			require.Nil(t, serve.Server)
		})

		t.Run("without handler", func(t *testing.T) {
			v.SetDefault("api.address", ":8090")
			serve := NewAPIServer(v, l, nil)
			require.Nil(t, serve.Server)
		})

		t.Run("should be ok", func(t *testing.T) {
			assert := require.New(t)
			v.SetDefault("api.address", ":8090")
			serve := NewAPIServer(v, l, testHTTPHandler(assert))
			assert.NotNil(serve.Server)
			assert.IsType(&mserv.HTTPServer{}, serve.Server)
		})
	})

	t.Run("check multi server", func(t *testing.T) {
		var (
			err     error
			assert  = require.New(t)
			servers = map[string]net.Listener{
				"pprof.address":   nil,
				"metrics.address": nil,
				"api.address":     nil,
			}
		)

		// Randomize ports:
		for name := range servers {
			servers[name], err = net.Listen("tcp", "127.0.0.1:0")
			assert.NoError(err)
			assert.NoError(servers[name].Close())

			v.SetDefault(name, servers[name].Addr().String())
		}

		mod := module.Module{
			{Constructor: func() *viper.Viper { return v }},
			{Constructor: func() logger.StdLogger { return l }},
			{Constructor: func() http.Handler { return testHTTPHandler(assert) }},

			{
				Constructor: func() http.Handler { return testHTTPHandler(assert) },
				Options:     []dig.ProvideOption{dig.Name("metric_handler")},
			},

			{
				Constructor: func() http.Handler { return testHTTPHandler(assert) },
				Options:     []dig.ProvideOption{dig.Name("profile_handler")},
			},
		}.Append(
			ServersModule,
		)

		assert.NoError(module.Provide(di, mod))

		err = di.Invoke(func(serve mserv.Server) {
			assert.IsType(&mserv.MultiServer{}, serve)

			serve.Start()
		})
		assert.NoError(err)

		for name, lis := range servers {
			{
				t.Logf("check for %q on %q", name, lis.Addr())

				resp, err := http.Get("http://" + lis.Addr().String() + "/test")
				assert.NoError(err)

				defer func() {
					err := resp.Body.Close()
					assert.NoError(err)
				}()

				data, err := ioutil.ReadAll(resp.Body)
				assert.NoError(err)

				assert.Equal(expectResult, data)
			}
		}
	})
}
