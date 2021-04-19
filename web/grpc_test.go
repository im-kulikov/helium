package web

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/im-kulikov/helium/group"
)

const listenSize = 256 * 1024

func TestGRPCService(t *testing.T) {
	t.Run("should set address and network", func(t *testing.T) {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		require.NoError(t, lis.Close())

		serve, err := NewGRPCService(
			grpc.NewServer(),
			GRPCName(testGRPCServe),
			GRPCWithLogger(zaptest.NewLogger(t)),
			GRPCListenNetwork(lis.Addr().Network()),
			GRPCListenAddress(lis.Addr().String()))
		require.NoError(t, err)

		s, ok := serve.(*gRPC)
		require.True(t, ok)
		require.Equal(t, lis.Addr().String(), s.address)
		require.Equal(t, lis.Addr().Network(), s.network)
	})

	t.Run("should fail on empty address", func(t *testing.T) {
		serve, err := NewGRPCService(grpc.NewServer())
		require.Nil(t, serve)
		require.EqualError(t, err, ErrEmptyGRPCAddress.Error())
	})

	t.Run("should fail on empty server", func(t *testing.T) {
		serve, err := NewGRPCService(nil)
		require.Nil(t, serve)
		require.EqualError(t, err, ErrEmptyGRPCServer.Error())
	})

	t.Run("should fail on Start and Stop", func(t *testing.T) {
		require.EqualError(t, (&gRPC{}).Start(nil), ErrEmptyGRPCServer.Error())
		require.Panics(t, func() {
			(&gRPC{}).Stop(nil)
		}, ErrEmptyGRPCServer.Error())
	})

	t.Run("should fail on net.Listen", func(t *testing.T) {
		srv, err := NewGRPCService(grpc.NewServer(), GRPCListenAddress("test:80"))
		require.Nil(t, srv)
		require.Error(t, err)
		require.Contains(t, err.Error(), "listen tcp: lookup test")
	})

	t.Run("should ignore ErrServerStopped", func(t *testing.T) {
		lis := bufconn.Listen(listenSize)
		serve, err := NewGRPCService(
			grpc.NewServer(),
			GRPCSkipErrors(),
			GRPCWithLogger(zap.L()),
			GRPCListener(lis))
		require.NoError(t, err)

		require.NotEmpty(t, serve.Name())
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		require.NoError(t, group.New().Add(serve.Start, serve.Stop).Run(ctx))
	})
}
