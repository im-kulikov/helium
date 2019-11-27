package web

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type fakeListener struct {
	startError error
	stopError  error
}

func (f fakeListener) ListenAndServe() error {
	return f.startError
}

func (f fakeListener) Shutdown(context.Context) error {
	return f.stopError
}

func TestListenerService(t *testing.T) {
	t.Run("should set network", func(t *testing.T) {
		serve, err := NewListener(
			&fakeListener{},
			ListenerIgnoreError(ErrEmptyListener),
			ListenerSkipErrors(),
			ListenerShutdownTimeout(time.Second))
		require.NoError(t, err)

		s, ok := serve.(*listener)
		require.True(t, ok)
		require.True(t, s.skipErrors)
		require.Equal(t, time.Second, s.shutdownTimeout)
		require.Equal(t, ErrEmptyListener, s.ignoreErrors[0])
	})

	t.Run("should fail on empty server", func(t *testing.T) {
		serve, err := NewListener(nil)
		require.Nil(t, serve)
		require.EqualError(t, err, ErrEmptyListener.Error())
	})

	t.Run("should fail on Start and Stop", func(t *testing.T) {
		require.EqualError(t, (&listener{}).Start(), ErrEmptyListener.Error())
		require.EqualError(t, (&listener{}).Stop(), ErrEmptyListener.Error())
	})

	t.Run("should successfully start and stop", func(t *testing.T) {
		require.NoError(t, (&listener{server: &fakeListener{}}).Start())
		require.NoError(t, (&listener{server: &fakeListener{}}).Stop())
	})

	t.Run("should skip errors", func(t *testing.T) {
		s := &fakeListener{stopError: errors.New("stopping")}
		serve, err := NewListener(s, ListenerSkipErrors())
		require.NoError(t, err)
		require.NoError(t, serve.Stop())
	})

	t.Run("should ignore errors", func(t *testing.T) {
		s := &fakeListener{stopError: ErrEmptyListener}
		serve, err := NewListener(s, ListenerIgnoreError(ErrEmptyListener))
		require.NoError(t, err)
		require.NoError(t, serve.Stop())
	})
}
