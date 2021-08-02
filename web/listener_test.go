package web

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/im-kulikov/helium/internal"
)

type (
	fakeListener struct {
		startError error
		stopError  error
	}
)

const (
	listenerTestName = "test-name"

	errStopping = internal.Error("stopping")
)

var _ Listener = (*fakeListener)(nil)

func (f fakeListener) ListenAndServe() error {
	return f.startError
}

func (f fakeListener) Shutdown(context.Context) error {
	return f.stopError
}

func TestListenerService(t *testing.T) {
	log := zap.NewNop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	t.Run("should be configured", func(t *testing.T) {
		serve, err := NewListener(
			&fakeListener{},
			ListenerIgnoreError(ErrEmptyListener),
			ListenerSkipErrors(),
			ListenerName(listenerTestName),
			ListenerWithLogger(nil), // should ignore
			ListenerWithLogger(log))
		require.NoError(t, err)

		s, ok := serve.(*listener)
		require.True(t, ok)
		require.True(t, s.skipErrors)
		require.Equal(t, log, s.logger)
		require.Equal(t, listenerTestName, s.Name())
		require.Equal(t, ErrEmptyListener, s.ignoreErrors[0])
	})

	t.Run("should fail on empty server", func(t *testing.T) {
		serve, err := NewListener(nil)
		require.Nil(t, serve)
		require.EqualError(t, err, ErrEmptyListener.Error())
	})

	t.Run("should error on Start and Stop", func(t *testing.T) {
		l := newTestLogger()

		require.EqualError(t, (&listener{logger: l.Logger}).Start(ctx), ErrEmptyListener.Error())
		l.Cleanup()

		(&listener{name: listenerTestName, logger: l.Logger}).Stop(ctx)
		require.NoError(t, l.Decode())

		require.Equal(t, l.Result.N, listenerTestName)
		require.Equal(t, l.Result.M, ErrEmptyListener.Error())
		require.Equal(t, l.Result.L, zapcore.ErrorLevel.CapitalString())
	})

	t.Run("should successfully start and stop", func(t *testing.T) {
		l := newTestLogger()

		require.NoError(t, (&listener{
			logger: l.Logger,
			server: &fakeListener{},
		}).Start(ctx))

		l.Cleanup()

		(&listener{
			name:   listenerTestName,
			logger: l.Logger,
			server: &fakeListener{stopError: ErrEmptyListener},
		}).Stop(ctx)
		require.NoError(t, l.Decode())

		require.Equal(t, l.Result.N, listenerTestName)
		require.Equal(t, l.Result.E, ErrEmptyListener.Error())
		require.Equal(t, l.Result.M, "could not stop listener")
		require.Equal(t, l.Result.L, zapcore.ErrorLevel.CapitalString())
	})

	t.Run("should skip errors", func(t *testing.T) {
		l := newTestLogger()
		lis := &fakeListener{stopError: errStopping}

		serve, err := NewListener(lis,
			ListenerSkipErrors(),
			ListenerWithLogger(l.Logger))
		require.NoError(t, err)

		serve.Stop(ctx)
		require.True(t, l.Empty())
	})

	t.Run("should ignore errors", func(t *testing.T) {
		l := newTestLogger()
		lis := &fakeListener{stopError: ErrEmptyListener}

		serve, err := NewListener(lis,
			ListenerIgnoreError(ErrEmptyListener),
			ListenerWithLogger(l.Logger))
		require.NoError(t, err)

		serve.Stop(ctx)
		require.True(t, l.Empty())
	})
}
