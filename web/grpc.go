package web

import (
	"context"
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/im-kulikov/helium/internal"
	"github.com/im-kulikov/helium/service"
)

type (
	gRPC struct {
		skipErrors bool
		name       string
		address    string
		network    string
		listener   net.Listener
		logger     *zap.Logger
		server     *grpc.Server
	}

	// GRPCOption allows changing default gRPC
	// service settings.
	GRPCOption func(g *gRPC)
)

const (
	// ErrEmptyGRPCServer is raised when called NewGRPCService
	// or gRPC methods with empty grpc.Server.
	ErrEmptyGRPCServer = internal.Error("empty gRPC server")

	// ErrEmptyGRPCAddress is raised when passed empty address to NewGRPCService.
	ErrEmptyGRPCAddress = internal.Error("empty gRPC address")
)

// GRPCSkipErrors allows to skip any errors.
func GRPCSkipErrors() GRPCOption {
	return func(g *gRPC) {
		g.skipErrors = true
	}
}

// GRPCName allows set name for the http-service.
func GRPCName(name string) GRPCOption {
	return func(s *gRPC) {
		s.name = name
	}
}

// GRPCListenAddress allows to change network for net.Listener.
func GRPCListenAddress(addr string) GRPCOption {
	return func(g *gRPC) {
		g.address = addr
	}
}

// GRPCListenNetwork allows to change default (tcp) network for net.Listener.
func GRPCListenNetwork(network string) GRPCOption {
	return func(g *gRPC) {
		g.network = network
	}
}

// GRPCListener allows to set custom net.Listener.
func GRPCListener(lis net.Listener) GRPCOption {
	return func(g *gRPC) {
		g.listener = lis
	}
}

// GRPCWithLogger changes default logger.
func GRPCWithLogger(l *zap.Logger) GRPCOption {
	return func(g *gRPC) {
		g.logger = l
	}
}

// NewGRPCService creates gRPC service with passed gRPC options.
// If something went wrong it returns an error.
func NewGRPCService(serve *grpc.Server, opts ...GRPCOption) (service.Service, error) {
	if serve == nil {
		return nil, ErrEmptyGRPCServer
	}

	s := &gRPC{
		server:     serve,
		network:    "tcp",
		skipErrors: false,
		logger:     zap.L(),
		name:       "unknown",
	}

	for i := range opts {
		opts[i](s)
	}

	if s.listener != nil {
		return s, nil
	}

	if s.address == "" {
		return nil, ErrEmptyGRPCAddress
	}

	var err error
	if s.listener, err = net.Listen(s.network, s.address); err != nil {
		return nil, s.catch(err)
	}

	return s, nil
}

// Name returns name of the service.
func (g *gRPC) Name() string {
	return fmt.Sprintf("gRPC(%s) %s", g.name, g.listener.Addr())
}

// Start tries to start gRPC service.
// If something went wrong it returns an error.
// If service could not start returns an error.
func (g *gRPC) Start(context.Context) error {
	if g.server == nil {
		return ErrEmptyGRPCServer
	}

	g.logger.Info("starting gRPC server",
		zap.String("name", g.name),
		zap.Stringer("address", g.listener.Addr()))

	return g.catch(g.server.Serve(g.listener))
}

// Stop tries to stop gRPC service.
func (g *gRPC) Stop(context.Context) {
	if g.server == nil {
		g.logger.Error("could not stop gRPC server",
			zap.String("name", g.name),
			zap.Error(ErrEmptyGRPCServer))

		return
	}

	g.server.GracefulStop()
}

func (g *gRPC) catch(err error) error {
	if g.skipErrors || errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}

	return err
}
