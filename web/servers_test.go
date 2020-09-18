package web

import (
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	gt "google.golang.org/grpc/test/grpc_testing"
)

type (
	testGRPC struct {
		gt.TestServiceServer
	}

	grpcResult struct {
		dig.Out

		Config string       `name:"grpc_config"`
		Server *grpc.Server `name:"grpc_server"`
	}
)

var (
	_ = ListenerSkipErrors
	_ = ListenerIgnoreError
	_ = ListenerShutdownTimeout
)

func (t testGRPC) EmptyCall(context.Context, *gt.Empty) (*gt.Empty, error) {
	return new(gt.Empty), nil
}

func (t testGRPC) UnaryCall(context.Context, *gt.SimpleRequest) (*gt.SimpleResponse, error) {
	return nil, status.Error(codes.AlreadyExists, codes.AlreadyExists.String())
}

var expectResult = []byte("OK")

func testHTTPHandler(assert *require.Assertions) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write(expectResult)
		assert.NoError(err)
	})
	return mux
}

func testGRPCServer(_ *require.Assertions) *grpc.Server {
	s := grpc.NewServer()
	gt.RegisterTestServiceServer(s, testGRPC{})
	return s
}

func TestServers(t *testing.T) {
	var (
		l  = zap.L()
		di = dig.New()
		v  = viper.New()
	)

	t.Run("gRPC default server", func(t *testing.T) {
		t.Run("should skip for disabled gRPC server", func(t *testing.T) {
			g := grpc.NewServer()
			v.Set("test-gRPC.disabled", true)
			res, err := newDefaultGRPCServer(grpcParams{
				Logger: l,
				Viper:  v,
				Key:    "test-gRPC",
				Server: g,
			})
			require.Empty(t, res)
			require.NoError(t, err)
			require.Empty(t, res)
		})

		t.Run("should skip for empty gRPC server", func(t *testing.T) {
			res, err := newDefaultGRPCServer(grpcParams{
				Logger: l,
				Viper:  v,
				Key:    "test-gRPC",
			})
			require.Empty(t, res)
			require.NoError(t, err)
			require.Empty(t, res)
		})

		t.Run("should fail for empty logger", func(t *testing.T) {
			res, err := newDefaultGRPCServer(grpcParams{})
			require.Empty(t, res)
			require.EqualError(t, err, ErrEmptyLogger.Error())
		})

		t.Run("should fail for empty viper", func(t *testing.T) {
			res, err := newDefaultGRPCServer(grpcParams{Logger: l, Key: "some-key"})
			require.Empty(t, res)
			require.NoError(t, err)
		})

		t.Run("should skip empty gRPC default server", func(t *testing.T) {
			res, err := newDefaultGRPCServer(grpcParams{Logger: l})
			require.Empty(t, res)
			require.NoError(t, err)
		})

		t.Run("should creates with passed config", func(t *testing.T) {
			v.Set("test_grpc.address", ":0")
			v.Set("test_grpc.network", "test")
			v.Set("test_grpc.skip_errors", true)

			res, err := newDefaultGRPCServer(grpcParams{
				Viper:  v,
				Logger: l,
				Key:    "test_grpc",
				Server: grpc.NewServer(),
			})
			require.NoError(t, err)

			serve, ok := res.Server.(*gRPC)
			require.True(t, ok)
			require.True(t, serve.skipErrors)
			require.Equal(t, serve.address, ":0")
			require.Equal(t, serve.network, "test")
		})
	})

	t.Run("check pprof server", func(t *testing.T) {
		t.Run("without logger", func(t *testing.T) {
			params := profileParams{Viper: v}
			serve, err := newProfileServer(params)
			require.EqualError(t, err, ErrEmptyLogger.Error())
			require.Nil(t, serve.Server)
		})

		t.Run("without config", func(t *testing.T) {
			params := profileParams{
				Viper:  v,
				Logger: l,
			}
			serve, err := newProfileServer(params)
			require.NoError(t, err)
			require.Nil(t, serve.Server)
		})

		t.Run("with config", func(t *testing.T) {
			v.SetDefault("pprof.address", ":6090")
			params := profileParams{
				Viper:  v,
				Logger: l,
			}
			serve, err := newProfileServer(params)
			require.NoError(t, err)
			require.NotNil(t, serve.Server)
			require.IsType(t, &httpService{}, serve.Server)
		})
	})

	t.Run("check metrics server", func(t *testing.T) {
		t.Run("without config", func(t *testing.T) {
			params := metricParams{
				Viper:  v,
				Logger: l,
			}
			serve, err := newMetricServer(params)
			require.NoError(t, err)
			require.Nil(t, serve.Server)
		})

		t.Run("with config", func(t *testing.T) {
			v.SetDefault("metrics.address", ":8090")
			params := metricParams{
				Viper:  v,
				Logger: l,
			}
			serve, err := newMetricServer(params)
			require.NoError(t, err)
			require.NotNil(t, serve.Server)
			require.IsType(t, &httpService{}, serve.Server)
		})
	})

	t.Run("empty viper or config key for http-server", func(t *testing.T) {
		is := require.New(t)

		v.SetDefault("test-api.disabled", true)
		z, err := zap.NewDevelopment()
		is.NoError(err)

		testHTTPHandler(is)

		t.Run("empty key", func(t *testing.T) {
			serve, err := NewHTTPServer(v, "", testHTTPHandler(is), z)
			require.NoError(t, err)
			require.Nil(t, serve.Server)
		})

		t.Run("empty viper", func(t *testing.T) {
			serve, err := NewHTTPServer(nil, "test-key", testHTTPHandler(is), z)
			require.NoError(t, err)
			require.Nil(t, serve.Server)
		})
	})

	t.Run("disabled http-server", func(t *testing.T) {
		is := require.New(t)

		v.SetDefault("test-api.disabled", true)
		z, err := zap.NewDevelopment()
		is.NoError(err)

		testHTTPHandler(is)

		serve, err := NewHTTPServer(v, "test-api", testHTTPHandler(is), z)
		is.NoError(err)
		is.Nil(serve.Server)
	})

	t.Run("api should be configured", func(t *testing.T) {
		is := require.New(t)

		v.SetDefault("another-api.address", "test")
		v.SetDefault("another-api.network", "test")
		v.SetDefault("another-api.skip_errors", true)

		z, err := zap.NewDevelopment()
		is.NoError(err)

		testHTTPHandler(is)

		serve, err := NewHTTPServer(v, "another-api", testHTTPHandler(is), z)
		is.NoError(err)

		s, ok := serve.Server.(*httpService)
		is.True(ok)
		is.True(s.skipErrors)
		is.Equal("test", s.address)
		is.Equal("test", s.network)
	})

	t.Run("check api server", func(t *testing.T) {
		t.Run("without config", func(t *testing.T) {
			serve, err := NewAPIServer(APIParams{Config: v, Logger: l})
			require.NoError(t, err)
			require.Nil(t, serve.Server)
		})

		t.Run("without logger", func(t *testing.T) {
			v.SetDefault("api.address", ":8090")
			serve, err := NewAPIServer(APIParams{})
			require.EqualError(t, err, ErrEmptyLogger.Error())
			require.Nil(t, serve.Server)
		})

		t.Run("without handler", func(t *testing.T) {
			v.SetDefault("api.address", ":8090")
			serve, err := NewAPIServer(APIParams{Config: v, Logger: l})
			require.NoError(t, err)
			require.Nil(t, serve.Server)
		})

		t.Run("should be ok", func(t *testing.T) {
			assert := require.New(t)
			v.SetDefault("api.address", ":8090")
			serve, err := NewAPIServer(APIParams{
				Config:  v,
				Logger:  l,
				Handler: testHTTPHandler(assert),
			})
			assert.NoError(err)
			assert.NotNil(serve.Server)
			assert.IsType(&httpService{}, serve.Server)
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
				"grpc.address":    nil,
			}
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Randomize ports:
		for name := range servers {
			servers[name], err = net.Listen("tcp", "127.0.0.1:0")
			assert.NoError(err)
			assert.NoError(servers[name].Close())
			v.SetDefault(name, servers[name].Addr().String())
		}

		mod := module.Module{
			{Constructor: func() *zap.Logger { return l }},
			{Constructor: func() *viper.Viper { return v }},
			{Constructor: func() http.Handler { return testHTTPHandler(assert) }},
			{Constructor: func() grpcResult {
				return grpcResult{
					Config: "grpc",
					Server: testGRPCServer(assert),
				}
			}},

			{
				Constructor: func() http.Handler { return testHTTPHandler(assert) },
				Options:     []dig.ProvideOption{dig.Name("metric_handler")},
			},

			{
				Constructor: func() http.Handler { return testHTTPHandler(assert) },
				Options:     []dig.ProvideOption{dig.Name("profile_handler")},
			},
		}.Append(
			DefaultServersModule,
		)

		assert.NoError(module.Provide(di, mod))

		err = di.Invoke(func(serve service.Service) {
			assert.NotNil(serve)
			assert.NoError(serve.Start(ctx))

			for name, lis := range servers {
				t.Run(name, func(t *testing.T) {
					t.Logf("check for %q on %q", name, lis.Addr())

					switch name {
					case "grpc.address":
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
						defer cancel()

						conn, err := grpc.DialContext(ctx, lis.Addr().String(),
							grpc.WithBlock(),
							grpc.WithInsecure())

						assert.NoError(err)

						cli := gt.NewTestServiceClient(conn)

						{ // EmptyCall
							res, err := cli.EmptyCall(ctx, &gt.Empty{})
							assert.NoError(err)
							assert.NotNil(res)
						}

						{ // UnaryCall
							res, err := cli.UnaryCall(ctx, &gt.SimpleRequest{})
							assert.Nil(res)
							assert.Error(err)

							st, ok := status.FromError(err)
							assert.True(ok)
							assert.Equal(codes.AlreadyExists, st.Code())
							assert.Equal(codes.AlreadyExists.String(), st.Message())
						}

					default:
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
				})
			}

			assert.NoError(serve.Stop())
		})
		assert.NoError(err)
	})
}
