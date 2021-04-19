package web

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/test/bufconn"

	"github.com/im-kulikov/helium/group"
)

func TestHTTPService(t *testing.T) {
	t.Run("should set network and address", func(t *testing.T) {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		require.NoError(t, lis.Close())

		serve, err := NewHTTPService(
			&http.Server{},
			HTTPSkipErrors(),
			HTTPName(apiServer),
			HTTPListener(nil),
			HTTPWithLogger(nil),
			HTTPListenAddress(lis.Addr().String()),
			HTTPListenNetwork(lis.Addr().Network()))
		require.NoError(t, err)

		s, ok := serve.(*httpService)
		require.True(t, ok)
		require.Equal(t, lis.Addr().String(), s.address)
		require.Equal(t, lis.Addr().Network(), s.network)
	})

	t.Run("should fail on empty address", func(t *testing.T) {
		serve, err := NewHTTPService(&http.Server{})
		require.Nil(t, serve)
		require.EqualError(t, err, ErrEmptyHTTPAddress.Error())
	})

	t.Run("should fail on empty server", func(t *testing.T) {
		serve, err := NewHTTPService(nil)
		require.Nil(t, serve)
		require.EqualError(t, err, ErrEmptyHTTPServer.Error())
	})

	t.Run("should fail on Start and Stop", func(t *testing.T) {
		require.EqualError(t, (&httpService{}).Start(context.Background()), ErrEmptyHTTPServer.Error())
		require.Panics(t, func() {
			(&httpService{}).Stop(context.Background())
		}, ErrEmptyHTTPServer.Error())
	})

	t.Run("should fail on net.Listen", func(t *testing.T) {
		srv, err := NewHTTPService(&http.Server{}, HTTPListenAddress("test:80"))
		require.Nil(t, srv)
		require.EqualError(t, err, "listen tcp: lookup test: no such host")
	})

	t.Run("should fail for serve", func(t *testing.T) {
		lis := bufconn.Listen(listenSize)
		s := &http.Server{}

		serve, err := NewHTTPService(s,
			HTTPListener(lis),
			HTTPWithLogger(zaptest.NewLogger(t)))
		require.NoError(t, err)
		require.NotEmpty(t, serve.Name())

		ctx, cancel := context.WithCancel(context.Background())
		require.NoError(t, group.New().
			Add(serve.Start, func(context.Context) { serve.Stop(ctx) }).
			Add(func(ctx context.Context) error {
				cancel()

				con, errConn := lis.Dial()
				if errConn != nil {
					return errConn
				}

				go func() {
					// emulate long query:
					time.Sleep(time.Second)

					_ = con.Close()
				}()

				return nil
			}, func(context.Context) {}).
			Run(ctx))
	})

	t.Run("should not fail for tls", func(t *testing.T) {
		lis := bufconn.Listen(listenSize)
		s := &http.Server{
			TLSConfig: &tls.Config{
				GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
					return nil, errors.New("test")
				},
			},
		}

		serve, err := NewHTTPService(s,
			HTTPListener(lis),
			HTTPWithLogger(zaptest.NewLogger(t)))
		require.NoError(t, err)
		require.NotEmpty(t, serve.Name())

		ctx, cancel := context.WithCancel(context.Background())
		require.NoError(t, group.New().
			Add(serve.Start, serve.Stop).
			Add(func(context.Context) error {
				cancel()

				return nil
			}, func(context.Context) {}).
			Run(ctx))
	})
}
