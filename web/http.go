package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/im-kulikov/helium/internal"
)

type (
	httpService struct {
		skipErrors      bool
		address         string
		network         string
		server          *http.Server
		shutdownTimeout time.Duration
	}

	// HTTPOption interface that allows
	// to change default http-service options.
	HTTPOption func(s *httpService)
)

const (
	// ErrEmptyHTTPServer is raised when called New
	// or httpService methods with empty http.Server.
	ErrEmptyHTTPServer = internal.Error("empty http server")

	// ErrEmptyHTTPAddress is raised when passed empty address to NewHTTPService.
	ErrEmptyHTTPAddress = internal.Error("empty http address")
)

// HTTPShutdownTimeout changes default shutdown timeout.
func HTTPShutdownTimeout(v time.Duration) HTTPOption {
	return func(s *httpService) {
		s.shutdownTimeout = v
	}
}

// HTTPListenNetwork allows to change default (tcp)
// network for net.Listener.
func HTTPListenNetwork(network string) HTTPOption {
	return func(s *httpService) {
		s.network = network
	}
}

// HTTPListenAddress allows to change network for net.Listener.
// By default it takes address from http.Server.
func HTTPListenAddress(address string) HTTPOption {
	return func(s *httpService) {
		s.address = address
	}
}

// HTTPSkipErrors allows to skip any errors
func HTTPSkipErrors() HTTPOption {
	return func(s *httpService) {
		s.skipErrors = true
	}
}

// NewHTTPService creates Service from http.Server and HTTPOption's.
func NewHTTPService(serve *http.Server, opts ...HTTPOption) (Service, error) {
	if serve == nil {
		return nil, ErrEmptyHTTPServer
	}

	s := &httpService{
		skipErrors:      false,
		server:          serve,
		network:         "tcp",
		address:         serve.Addr,
		shutdownTimeout: time.Second * 30,
	}

	for i := range opts {
		opts[i](s)
	}

	if s.address == "" {
		return nil, ErrEmptyHTTPAddress
	}

	return s, nil
}

// Start runs http.Server and returns error
// if something went wrong.
func (s *httpService) Start() error {
	var (
		err error
		lis net.Listener
	)

	if s.server == nil {
		return s.catch(ErrEmptyHTTPServer)
	} else if lis, err = net.Listen(s.network, s.address); err != nil {
		return s.catch(err)
	}

	go func() {
		var err error

		switch {
		case s.server.TLSConfig == nil:
			err = s.server.Serve(lis)
		default:
			// provide cert and key from TLSConfig
			err = s.server.ServeTLS(lis, "", "")
		}

		// ignores known error
		if err = s.catch(err); err != nil {
			fmt.Printf("could not start http.Server: %v\n", err)
			fatal(2)
		}
	}()

	return nil
}

// Stop tries to stop http.Server and returns error
// if something went wrong.
func (s *httpService) Stop() error {
	if s.server == nil {
		return s.catch(ErrEmptyHTTPServer)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.catch(s.server.Shutdown(ctx))
}

func (s *httpService) catch(err error) error {
	if s.skipErrors || err == http.ErrServerClosed {
		return nil
	}

	return err
}
