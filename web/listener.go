package web

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/im-kulikov/helium/internal"
	"github.com/im-kulikov/helium/service"
)

type (
	// Listener service interface.
	Listener interface {
		ListenAndServe() error
		Shutdown(context.Context) error
	}

	listener struct {
		logger *zap.Logger

		name         string
		skipErrors   bool
		ignoreErrors []error
		server       Listener
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

// ListenerName allows changing the default listener name.
func ListenerName(v string) ListenerOption {
	return func(l *listener) {
		l.name = v
	}
}

func ListenerWithLogger(log *zap.Logger) ListenerOption {
	return func(l *listener) {
		if log == nil {
			return
		}

		l.logger = log
	}
}

// NewListener creates new Listener service and applies passed options to it.
func NewListener(lis Listener, opts ...ListenerOption) (service.Service, error) {
	if lis == nil {
		return nil, ErrEmptyListener
	}

	s := &listener{
		server:     lis,
		skipErrors: false,

		// Default name
		name: fmt.Sprintf("listener %T", lis),
	}

	for i := range opts {
		opts[i](s)
	}

	return s, nil
}

// Name returns name of the service.
func (l *listener) Name() string { return l.name }

// Start tries to start the Listener and returns an error
// if the Listener is empty. If something went wrong and
// errors not ignored should panic.
func (l *listener) Start(context.Context) error {
	if l.server == nil {
		return l.catch(ErrEmptyListener)
	}

	return l.catch(l.server.ListenAndServe())
}

// Stop tries to stop the Listener and returns an error
// if something went wrong. Ignores errors that were passed
// by options and if used skip errors.
func (l *listener) Stop(ctx context.Context) {
	if l.server == nil {
		panic(ErrEmptyListener)
	}

	if err := l.catch(l.server.Shutdown(ctx)); err != nil {
		l.logger.Error("could not stop listener",
			zap.String("name", l.name),
			zap.Error(err))
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
