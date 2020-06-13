package web

import (
	"context"
	"errors"
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type fakeService struct {
	startError error
	stopError  error
}

func (f fakeService) Name() string { return "fake" }

func (f fakeService) Start(_ context.Context) error {
	return f.startError
}

func (f fakeService) Stop() error {
	return f.stopError
}

func TestMultiService(t *testing.T) {
	t.Run("fail on empty logger", func(t *testing.T) {
		svc, err := New(nil)
		require.Nil(t, svc)
		require.EqualError(t, err, ErrEmptyLogger.Error())
	})

	t.Run("fail on empty services", func(t *testing.T) {
		svc, err := New(zap.L(), nil, nil)
		require.Nil(t, svc)
		require.EqualError(t, err, ErrEmptyServices.Error())
	})

	t.Run("should fail on start and return first error", func(t *testing.T) {
		svc, err := New(zap.L(),
			&fakeService{startError: ErrEmptyServices},
			&fakeService{startError: ErrEmptyLogger})
		require.NoError(t, err)
		require.EqualError(t, svc.Start(context.Background()), ErrEmptyServices.Error())
	})

	t.Run("should fail on stop and return last error", func(t *testing.T) {
		l, err := zap.NewDevelopment()
		require.NoError(t, err)

		svc, err := New(l,
			&fakeService{stopError: ErrEmptyServices},
			&fakeService{stopError: ErrEmptyLogger})
		require.NoError(t, err)
		require.NotEmpty(t, svc.Name())
		require.EqualError(t, svc.Stop(), ErrEmptyLogger.Error())
	})
}

func canceler(t *testing.T, cancel func()) func(int) {
	return func(code int) {
		require.Equal(t, 2, code)
		cancel()
	}
}

func Test_ShouldFailInGoroutine(t *testing.T) {
	{ // HTTP server:
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		require.NoError(t, lis.Close())

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		fatal = canceler(t, cancel)

		serve := &http.Server{}
		monkey.PatchInstanceMethod(reflect.TypeOf(serve), "Serve", func(*http.Server, net.Listener) error {
			return errors.New("done")
		})

		s, err := NewHTTPService(serve, HTTPListenAddress(lis.Addr().String()))
		require.NoError(t, err)
		require.NoError(t, s.Start(ctx))

		<-ctx.Done()
		require.EqualError(t, ctx.Err(), context.Canceled.Error())
	}

	{ // gRPC server:
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		require.NoError(t, lis.Close())

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		fatal = canceler(t, cancel)

		serve := grpc.NewServer()
		monkey.PatchInstanceMethod(reflect.TypeOf(serve), "Serve", func(*grpc.Server, net.Listener) error {
			return errors.New("done")
		})

		s, err := NewGRPCService(serve, GRPCListenAddress(lis.Addr().String()))
		require.NoError(t, err)
		require.NoError(t, s.Start(ctx))

		<-ctx.Done()
		require.EqualError(t, ctx.Err(), context.Canceled.Error())
	}

	{ // Listener
		s := &fakeListener{startError: errors.New("done")}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		fatal = canceler(t, cancel)

		serve, err := NewListener(s)
		require.NoError(t, err)
		require.NoError(t, serve.Start(ctx))

		<-ctx.Done()
		require.EqualError(t, ctx.Err(), context.Canceled.Error())
	}
}
