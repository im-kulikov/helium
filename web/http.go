package web

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"

	"github.com/im-kulikov/helium/internal"
	"github.com/im-kulikov/helium/service"
)

type (
	httpService struct {
		logger *zap.Logger

		skipErrors bool
		name       string
		address    string
		network    string
		listener   net.Listener
		server     *http.Server
	}

	// HTTPOption interface that allows
	// to change default http-service options.
	HTTPOption func(s *httpService)
)

const (
	// ErrEmptyHTTPServer is raised when called New or httpService methods with empty http.Server.
	ErrEmptyHTTPServer = internal.Error("empty http server")

	// ErrEmptyHTTPAddress is raised when passed empty address to NewHTTPService.
	ErrEmptyHTTPAddress = internal.Error("empty http address")
)

// HTTPName allows set name for the http-service.
func HTTPName(name string) HTTPOption {
	return func(s *httpService) {
		s.name = name
	}
}

// HTTPListenNetwork allows changing default (tcp) network for net.Listener.
func HTTPListenNetwork(network string) HTTPOption {
	return func(s *httpService) {
		s.network = network
	}
}

// HTTPListenAddress allows changing network for net.Listener.
// By default, it takes address from http.Server.
func HTTPListenAddress(address string) HTTPOption {
	return func(s *httpService) {
		s.address = address
	}
}

// HTTPSkipErrors allows to skip any errors.
func HTTPSkipErrors() HTTPOption {
	return func(s *httpService) {
		s.skipErrors = true
	}
}

// HTTPListener allows to set custom net.Listener.
func HTTPListener(lis net.Listener) HTTPOption {
	return func(s *httpService) {
		if lis == nil {
			return
		}

		s.listener = lis
	}
}

// HTTPWithLogger allows to set logger.
func HTTPWithLogger(l *zap.Logger) HTTPOption {
	return func(s *httpService) {
		if l == nil {
			return
		}

		s.logger = l
	}
}

// NewHTTPService creates Service from http.Server and HTTPOption's.
func NewHTTPService(serve *http.Server, opts ...HTTPOption) (service.Service, error) {
	if serve == nil {
		return nil, ErrEmptyHTTPServer
	}

	s := &httpService{
		logger: zap.NewNop(),

		skipErrors: false,
		server:     serve,
		network:    "tcp",
	}

	for i := range opts {
		opts[i](s)
	}

	if s.listener != nil {
		return s, nil
	}

	if s.address == "" {
		return nil, ErrEmptyHTTPAddress
	}

	var err error
	if s.listener, err = net.Listen(s.network, s.address); err != nil {
		return nil, s.catch(err)
	}

	return s, nil
}

// Name returns name of the service.
func (s *httpService) Name() string {
	return fmt.Sprintf("http(%s) %s", s.name, s.listener.Addr())
}

// Start runs http.Server and returns error
// if something went wrong.
func (s *httpService) Start(context.Context) error {
	if s.server == nil {
		return ErrEmptyHTTPServer
	}

	switch {
	case s.server.TLSConfig == nil:
		return s.catch(s.server.Serve(s.listener))
	default:
		// provide cert and key from TLSConfig
		return s.catch(s.server.ServeTLS(s.listener, "", ""))
	}
}

// Stop tries to stop http.Server and returns error
// if something went wrong.
func (s *httpService) Stop(ctx context.Context) {
	if s.server == nil {
		s.logger.Error("could not stop http.Server",
			zap.String("name", s.name),
			zap.Error(ErrEmptyHTTPServer))

		return
	}

	if err := s.catch(s.server.Shutdown(ctx)); err != nil {
		s.logger.Error("could not stop http.Server",
			zap.String("name", s.name),
			zap.Error(err))
	}
}

func (s *httpService) catch(err error) error {
	if s.skipErrors || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
