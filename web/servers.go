package web

import (
	"net/http"
	"net/http/pprof"

	"github.com/im-kulikov/helium/module"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	// APIParams struct
	APIParams struct {
		dig.In

		Config  *viper.Viper
		Logger  *zap.Logger
		Handler http.Handler `optional:"true"`
	}

	// MultiServerParams struct
	MultiServerParams struct {
		dig.In

		Logger  *zap.Logger
		Servers []Service `group:"services"`
	}

	// ServerResult struct
	ServerResult struct {
		dig.Out

		Server Service `group:"services"`
	}

	profileParams struct {
		dig.In

		Logger  *zap.Logger
		Viper   *viper.Viper
		Handler http.Handler `name:"profile_handler" optional:"true"`
	}

	metricParams struct {
		dig.In

		Logger  *zap.Logger
		Viper   *viper.Viper
		Handler http.Handler `name:"metric_handler" optional:"true"`
	}

	grpcParams struct {
		dig.In

		Logger *zap.Logger
		Viper  *viper.Viper
		Key    string       `name:"grpc_config" optional:"true"`
		Server *grpc.Server `name:"grpc_server" optional:"true"`
	}
)

var (
	// ServersModule of web base structs
	ServersModule = module.Module{
		{Constructor: newDefaultGRPCServer},
		{Constructor: newProfileServer},
		{Constructor: newMetricServer},
		{Constructor: NewAPIServer},
		{Constructor: NewMultiServer},
	}
)

// NewMultiServer returns new multi servers group
func NewMultiServer(p MultiServerParams) (Service, error) { return New(p.Logger, p.Servers...) }

func newProfileServer(p profileParams) (ServerResult, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	if p.Handler != nil {
		mux.Handle("/", p.Handler)
	}
	return NewHTTPServer(p.Viper, "pprof", mux, p.Logger)
}

func newMetricServer(p metricParams) (ServerResult, error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	if p.Handler != nil {
		mux.Handle("/", p.Handler)
	}
	return NewHTTPServer(p.Viper, "metrics", mux, p.Logger)
}

func newDefaultGRPCServer(p grpcParams) (ServerResult, error) {
	return NewGRPCServer(p.Viper, p.Key, p.Server, p.Logger)
}

// NewAPIServer creates api server by http.Handler from DI container
func NewAPIServer(p APIParams) (ServerResult, error) {
	return NewHTTPServer(p.Config, "api", p.Handler, p.Logger)
}

// NewGRPCServer creates gRPC server that will be embedded info multi-server.
func NewGRPCServer(v *viper.Viper, key string, s *grpc.Server, l *zap.Logger) (ServerResult, error) {
	switch {
	case l == nil:
		return ServerResult{}, ErrEmptyLogger
	case key == "" || v == nil:
		l.Info("Empty config or key for gRPC server, skip")
		return ServerResult{}, nil
	case s == nil:
		l.Info("Empty server, skip",
			zap.String("name", key))
		return ServerResult{}, nil
	case v.IsSet(key + ".disabled"):
		l.Info("Disabled, skip",
			zap.String("name", key))
		return ServerResult{}, nil
	}

	options := []GRPCOption{
		GRPCListenAddress(v.GetString(key + ".address")),
		GRPCShutdownTimeout(v.GetDuration(key + ".shutdown_timeout")),
	}

	if v.GetBool(key + ".skip_errors") {
		options = append(options, GRPCSkipErrors())
	}

	if v.IsSet(key + ".network") {
		options = append(options, GRPCListenNetwork(v.GetString(key+".network")))
	}

	serve, err := NewGRPCService(s, options...)

	l.Info("Creates gRPC server",
		zap.String("name", key),
		zap.String("address", v.GetString(key+".address")))

	return ServerResult{Server: serve}, err
}

// NewHTTPServer creates http-server that will be embedded into multi-server
func NewHTTPServer(v *viper.Viper, key string, h http.Handler, l *zap.Logger) (ServerResult, error) {
	switch {
	case l == nil:
		return ServerResult{}, ErrEmptyLogger
	case key == "" || v == nil:
		l.Info("Empty config or key for http server, skip")
		return ServerResult{}, nil
	case h == nil:
		l.Info("Empty handler, skip",
			zap.String("name", key))
		return ServerResult{}, nil
	case v.GetBool(key + ".disabled"):
		l.Info("Server disabled",
			zap.String("name", key))
		return ServerResult{}, nil
	case !v.IsSet(key + ".address"):
		l.Info("Empty bind address, skip",
			zap.String("name", key))
		return ServerResult{}, nil
	}

	options := []HTTPOption{
		HTTPListenAddress(v.GetString(key + ".address")),
		HTTPShutdownTimeout(v.GetDuration(key + ".shutdown_timeout")),
	}

	if v.IsSet(key + ".network") {
		options = append(options, HTTPListenNetwork(v.GetString(key+".network")))
	}

	if v.IsSet(key + ".skip_errors") {
		options = append(options, HTTPSkipErrors())
	}

	serve, err := NewHTTPService(
		&http.Server{
			Handler:           h,
			Addr:              v.GetString(key + ".address"),
			ReadTimeout:       v.GetDuration(key + ".read_timeout"),
			ReadHeaderTimeout: v.GetDuration(key + ".read_header_timeout"),
			WriteTimeout:      v.GetDuration(key + ".write_timeout"),
			IdleTimeout:       v.GetDuration(key + ".idle_timeout"),
			MaxHeaderBytes:    v.GetInt(key + ".max_header_bytes"),
		},
		options...,
	)

	l.Info("Creates http server",
		zap.String("name", key),
		zap.String("address", v.GetString(key+".address")))

	return ServerResult{Server: serve}, err
}
