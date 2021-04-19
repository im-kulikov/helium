package web

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type (
	fakeListener struct {
		startError error
		stopError  error
	}
)

const listenerTestName = "test-name"

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

	t.Run("should fail on Start and Stop", func(t *testing.T) {
		require.EqualError(t, (&listener{}).Start(ctx), ErrEmptyListener.Error())
		require.Panics(t, func() {
			(&listener{logger: log}).Stop(ctx)
		}, ErrEmptyListener.Error())
	})

	t.Run("should successfully start and stop", func(t *testing.T) {
		require.NoError(t, (&listener{server: &fakeListener{}}).Start(ctx))
		require.NotPanics(t, func() {
			(&listener{
				logger: log,
				server: &fakeListener{stopError: ErrEmptyLogger},
			}).Stop(ctx)
		})
	})

	t.Run("should skip errors", func(t *testing.T) {
		s := &fakeListener{stopError: errors.New("stopping")}
		serve, err := NewListener(s, ListenerSkipErrors())
		require.NoError(t, err)
		require.NotPanics(t, func() {
			serve.Stop(ctx)
		})
	})

	t.Run("should ignore errors", func(t *testing.T) {
		s := &fakeListener{stopError: ErrEmptyListener}
		serve, err := NewListener(s, ListenerIgnoreError(ErrEmptyListener))
		require.NoError(t, err)
		require.NotPanics(t, func() {
			serve.Stop(ctx)
		})
	})
}
