package web

import (
	"bytes"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	gt "google.golang.org/grpc/test/grpc_testing"

	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
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

	testMultiParams struct {
		dig.In

		Service service.Group
	}
)

const (
	testHTTPServe = "test-http"
	testGRPCServe = "test-grpc"
)

// One empty request followed by one empty response.
func (t testGRPC) EmptyCall(context.Context, *gt.Empty) (*gt.Empty, error) {
	return new(gt.Empty), nil
}

// One request followed by one response.
// The server returns the client payload as-is.
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
	gt.RegisterTestServiceServer(s, new(testGRPC))

	return s
}

func TestServers(t *testing.T) {
	var (
		l = zap.L()
		v = viper.New()
	)

	t.Run("gRPC default server", func(t *testing.T) {
		t.Run("should skip for empty gRPC server", func(t *testing.T) {
			res, err := newDefaultGRPCServer(grpcParams{Logger: l, Viper: v, Key: testGRPCServe})
			require.Empty(t, res)
			require.NoError(t, err)
			require.Empty(t, res)
		})

		t.Run("should skip for disabled gRPC server", func(t *testing.T) {
			v.Set("disabled-grpc.disabled", true)

			res, err := newDefaultGRPCServer(grpcParams{
				Logger: l,
				Viper:  v,
				Key:    "disabled-grpc",
				Server: grpc.NewServer(),
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
			res, err := newDefaultGRPCServer(grpcParams{Logger: zaptest.NewLogger(t), Key: testGRPCServe})
			require.Empty(t, res)
			require.NoError(t, err)
		})

		t.Run("should skip empty gRPC default server", func(t *testing.T) {
			res, err := newDefaultGRPCServer(grpcParams{Logger: zaptest.NewLogger(t)})
			require.Empty(t, res)
			require.NoError(t, err)
		})

		t.Run("should creates with passed config", func(t *testing.T) {
			lis := bufconn.Listen(listenSize)
			defer require.NoError(t, lis.Close())
			v.Set(testGRPCServe+".address", ":0")
			v.Set(testGRPCServe+".network", "test")
			v.Set(testGRPCServe+".disabled", false)
			v.Set(testGRPCServe+".skip_errors", true)

			res, err := newDefaultGRPCServer(grpcParams{
				Viper:    v,
				Listener: lis,
				Key:      testGRPCServe,
				Name:     testGRPCServe,
				Server:   grpc.NewServer(),
				Logger:   zaptest.NewLogger(t),
			})
			require.NoError(t, err)

			serve, ok := res.Server.(*gRPC)
			require.True(t, ok)
			require.True(t, serve.skipErrors)
			require.Equal(t, lis, serve.listener)
			require.Equal(t, serve.address, ":0")
			require.Equal(t, serve.network, "test")
		})
	})

	t.Run("empty viper or config key for http-server", func(t *testing.T) {
		is := require.New(t)

		v.SetDefault("test-api.disabled", true)

		testHTTPHandler(is)

		t.Run("empty key", func(t *testing.T) {
			serve, err := NewHTTPServer(HTTPParams{Logger: zaptest.NewLogger(t)})
			require.NoError(t, err)
			require.Nil(t, serve.Server)
		})

		t.Run("empty viper", func(t *testing.T) {
			serve, err := NewHTTPServer(HTTPParams{Logger: zaptest.NewLogger(t), Key: testHTTPServe})
			require.NoError(t, err)
			require.Nil(t, serve.Server)
		})

		t.Run("empty http-address", func(t *testing.T) {
			serve, err := NewHTTPServer(HTTPParams{
				Key:     testHTTPServe,
				Config:  viper.New(),
				Handler: http.NewServeMux(),
				Logger:  zaptest.NewLogger(t),
			})
			require.Nil(t, serve.Server)
			require.EqualError(t, err, ErrEmptyHTTPAddress.Error())
		})
	})

	t.Run("disabled http-server", func(t *testing.T) {
		is := require.New(t)

		v.SetDefault("test-api.disabled", true)

		testHTTPHandler(is)

		serve, err := NewHTTPServer(HTTPParams{Logger: zaptest.NewLogger(t)})
		is.NoError(err)
		is.Nil(serve.Server)
	})

	t.Run("api should be configured", func(t *testing.T) {
		is := require.New(t)

		name := "another-api"
		v.SetDefault(name+".skip_errors", true)

		z, err := zap.NewDevelopment()
		is.NoError(err)

		lis := bufconn.Listen(listenSize)

		serve, err := NewHTTPServer(HTTPParams{
			Config:   v,
			Logger:   z,
			Name:     name,
			Key:      name,
			Listener: lis,
			Handler:  testHTTPHandler(is),
		})
		is.NoError(err)

		s, ok := serve.Server.(*httpService)
		is.True(ok)
		is.True(s.skipErrors)
		is.Equal(lis, s.listener)
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
			listen := bufconn.Listen(listenSize)

			serve, err := NewAPIServer(APIParams{
				Config:   v,
				Logger:   l,
				Listener: listen,
				Handler:  testHTTPHandler(assert),
			})
			assert.NoError(err)
			assert.NotNil(serve.Server)
			assert.IsType(&httpService{}, serve.Server)
		})
	})

	t.Run("check multi server", func(t *testing.T) {
		var (
			cnr = dig.New()
			cfg = viper.New()
			log = zaptest.NewLogger(t)

			assert = require.New(t)
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cfg.SetDefault(testHTTPServe+".disabled", true)
		cfg.SetDefault(apiServer+".network", "tcp")
		cfg.SetDefault(apiServer+".address", "127.0.0.1")

		cfg.SetDefault(apiServer+".read_timeout", time.Second)
		cfg.SetDefault(apiServer+".idle_timeout", time.Second)
		cfg.SetDefault(apiServer+".write_timeout", time.Second)
		cfg.SetDefault(apiServer+".read_header_timeout", time.Second)
		cfg.SetDefault(apiServer+".max_header_bytes", math.MaxInt32)

		OpsDefaults(cfg)

		listeners := map[string]*bufconn.Listener{
			apiServer:  bufconn.Listen(listenSize),
			gRPCServer: bufconn.Listen(listenSize),
		}

		mod := module.Module{
			{Constructor: func() *zap.Logger { return log }},
			{Constructor: func() *viper.Viper { return cfg }},
			{Constructor: func() grpcResult {
				return grpcResult{
					Config: gRPCServer,
					Server: testGRPCServer(assert),
				}
			}},
			{
				Constructor: func() (ServerResult, error) {
					return NewHTTPServer(HTTPParams{
						Config:  cfg,
						Logger:  log,
						Name:    testHTTPServe,
						Key:     testHTTPServe,
						Handler: testHTTPHandler(assert),
					})
				},
			},
			{Constructor: func() http.Handler { return testHTTPHandler(assert) }},
			{
				Constructor: func() http.Handler { return testHTTPHandler(assert) },
				Options:     []dig.ProvideOption{dig.Name("pprof_handler")},
			},
			{
				Constructor: func() http.Handler { return testHTTPHandler(assert) },
				Options:     []dig.ProvideOption{dig.Name("metric_handler")},
			},
		}.Append(DefaultServersModule, service.Module)

		for item := range listeners {
			srv := item
			lis := listeners[srv]

			mod = mod.Append(module.Module{
				{
					Constructor: func() net.Listener { return lis },
					Options:     []dig.ProvideOption{dig.Name(srv + "_listener")},
				},
			})
		}

		assert.NoError(module.Provide(cnr, mod))

		buf := new(bytes.Buffer)
		require.NoError(t, dig.Visualize(cnr, buf))
		defer func() {
			if t.Failed() {
				t.Logf("\n%s", buf.String())
			}
		}()

		assert.NoError(cnr.Invoke(func(p testMultiParams) {
			assert.NotEmpty(p.Service)

			done := make(chan struct{})
			start := make(chan struct{})

			go func() {
				t.Helper()

				<-start
				assert.NoError(p.Service.Run(ctx))
				close(done)
			}()

			close(start)
			time.Sleep(time.Millisecond * 10)

			wg := new(sync.WaitGroup)
			wg.Add(len(listeners))

			for item := range listeners {
				srv := item
				lis := listeners[item]
				t.Run(srv, func(t *testing.T) {
					defer func() {
						t.Logf("done for %s", srv)
						wg.Done()
					}()

					switch srv {
					case gRPCServer:
						conn, err := grpc.DialContext(ctx, lis.Addr().String(),
							grpc.WithBlock(),
							grpc.WithTransportCredentials(insecure.NewCredentials()),
							grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
								return lis.Dial()
							}))

						require.NoError(t, err)

						cli := gt.NewTestServiceClient(conn)

						{ // EmptyCall
							res, err := cli.EmptyCall(ctx, &gt.Empty{})
							require.NoError(t, err)
							require.NotNil(t, res)
						}

						{ // UnaryCall
							res, err := cli.UnaryCall(ctx, &gt.SimpleRequest{})
							require.Nil(t, res)
							require.Error(t, err)

							st, ok := status.FromError(err)
							require.True(t, ok)
							require.Equal(t, codes.AlreadyExists, st.Code())
							require.Equal(t, codes.AlreadyExists.String(), st.Message())
						}

					default:
						client := &http.Client{
							Transport: &http.Transport{
								DialContext: func(context.Context, string, string) (net.Conn, error) {
									return lis.Dial()
								},
							},
						}

						// nolint:noctx
						resp, err := client.Get("http://" + lis.Addr().String() + "/test")
						require.NoError(t, err)

						data, err := ioutil.ReadAll(resp.Body)
						require.NoError(t, err)

						require.Equal(t, expectResult, data)
						require.NoError(t, resp.Body.Close())
					}
				})
			}

			wg.Wait()
			cancel()
			<-done
		}))
	})
}
