package web

import (
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"net"
	"testing"
)

func TestGRPCService(t *testing.T) {
	t.Run("should set network", func(t *testing.T) {
		serve, err := NewGRPCService(
			grpc.NewServer(),
			GRPCSkipErrors(),
			GRPCListenAddress(":8080"),
			GRPCListenNetwork("test"))
		require.NoError(t, err)

		s, ok := serve.(*gRPC)
		require.True(t, ok)
		require.Equal(t, "test", s.network)
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
		require.EqualError(t, (&gRPC{}).Start(), ErrEmptyGRPCServer.Error())
		require.EqualError(t, (&gRPC{}).Stop(), ErrEmptyGRPCServer.Error())
	})

	t.Run("should fail on net.Listen", func(t *testing.T) {
		require.EqualError(t, (&gRPC{server: grpc.NewServer()}).Start(), "listen: unknown network ")
	})

	t.Run("shoud ignore ErrServerStopped", func(t *testing.T) {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		require.NoError(t, lis.Close())

		serve, err := NewGRPCService(
			grpc.NewServer(),
			GRPCSkipErrors(),
			GRPCListenAddress(lis.Addr().String()))
		require.NoError(t, err)

		require.NoError(t, serve.Stop())
		require.NoError(t, serve.Start())
	})
}
