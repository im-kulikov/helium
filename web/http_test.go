package web

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPService(t *testing.T) {
	t.Run("should set network", func(t *testing.T) {
		serve, err := NewHTTPService(
			&http.Server{},
			HTTPSkipErrors(),
			HTTPListenAddress(":8080"),
			HTTPListenNetwork("test"))
		require.NoError(t, err)

		s, ok := serve.(*httpService)
		require.True(t, ok)
		require.Equal(t, "test", s.network)
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
		require.EqualError(t, (&httpService{}).Stop(), ErrEmptyHTTPServer.Error())
	})

	t.Run("should fail on net.Listen", func(t *testing.T) {
		require.EqualError(t, (&httpService{server: &http.Server{}}).Start(context.Background()), "listen: unknown network ")
	})

	t.Run("should not fail for tls", func(t *testing.T) {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		require.NoError(t, lis.Close())

		s := &http.Server{
			TLSConfig: &tls.Config{
				GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
					return nil, errors.New("test")
				},
			},
		}

		serve, err := NewHTTPService(s, HTTPListenAddress(lis.Addr().String()))
		require.NoError(t, err)
		require.NotEmpty(t, serve.Name())
		require.NoError(t, serve.Start(context.Background()))
	})
}
