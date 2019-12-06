package web

import (
	"context"
	"fmt"
	"time"

	"github.com/im-kulikov/helium/internal"
)

type (
	// Listener service interface.
	Listener interface {
		ListenAndServe() error
		Shutdown(context.Context) error
	}

	listener struct {
		skipErrors      bool
		ignoreErrors    []error
		server          Listener
		shutdownTimeout time.Duration
	}

	// ListenerOption options that allows to change
	// default settings for Listener service.
	ListenerOption func(l *listener)
)

// ErrEmptyListener is raised when an empty Listener
// used in the NewListener function or Listener methods.
const ErrEmptyListener = internal.Error("empty listener")

// ListenerSkipErrors allows for ignoring all raised errors.
func ListenerSkipErrors() ListenerOption {
	return func(l *listener) {
		l.skipErrors = true
	}
}

// ListenerIgnoreError allows for ignoring all passed errors.
func ListenerIgnoreError(errors ...error) ListenerOption {
	return func(l *listener) {
		l.ignoreErrors = errors
	}
}

// ListenerShutdownTimeout allows changing the default shutdown timeout.
func ListenerShutdownTimeout(v time.Duration) ListenerOption {
	return func(l *listener) {
		l.shutdownTimeout = v
	}
}

// NewListener creates new Listener service and applies passed options to it.
func NewListener(lis Listener, opts ...ListenerOption) (Service, error) {
	if lis == nil {
		return nil, ErrEmptyListener
	}

	s := &listener{
		server:          lis,
		skipErrors:      false,
		shutdownTimeout: time.Second * 30,
	}

	for i := range opts {
		opts[i](s)
	}

	return s, nil
}

// Start tries to start the Listener and returns an error
// if the Listener is empty. If something went wrong and
// errors not ignored should panic.
func (l *listener) Start() error {
	if l.server == nil {
		return l.catch(ErrEmptyListener)
	}

	go func() {
		if err := l.catch(l.server.ListenAndServe()); err != nil {
			fmt.Printf("could not start Listener: %v\n", err)
			fatal(2)
		}
	}()

	return nil
}

// Stop tries to stop the Listener and returns an error
// if something went wrong. Ignores errors that were passed
// by options and if used skip errors.
func (l *listener) Stop() error {
	if l.server == nil {
		return l.catch(ErrEmptyListener)
	}

	ctx, cancel := context.WithTimeout(context.Background(), l.shutdownTimeout)
	defer cancel()

	ch := make(chan error, 1)

	go func() {
		ch <- l.catch(l.server.Shutdown(ctx))
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-ch:
		return err
	}
}

func (l *listener) catch(err error) error {
	if l.skipErrors {
		return nil
	}

	for i := range l.ignoreErrors {
		if l.ignoreErrors[i] == err {
			return nil
		}
	}

	return err
}
