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

		Viper  *viper.Viper
		Key    string       `name:"grpc_config" optional:"true"`
		Server *grpc.Server `name:"grpc_server" optional:"true"`
	}
)

var (
	// ServersModule of web base structs
	ServersModule = module.Module{
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
	if p.Server == nil || p.Viper.IsSet(p.Key+".disabled") {
		return ServerResult{}, nil
	}

	options := []GRPCOption{
		GRPCListenAddress(p.Viper.GetString(p.Key + ".address")),
		GRPCShutdownTimeout(p.Viper.GetDuration(p.Key + ".shutdown_timeout")),
	}

	if p.Viper.GetBool(p.Key + ".skip_errors") {
		options = append(options, GRPCSkipErrors())
	}

	if p.Viper.IsSet(p.Key + ".network") {
		options = append(options, GRPCListenNetwork(p.Viper.GetString(p.Key+".network")))
	}

	serve, err := NewGRPCService(p.Server, options...)

	return ServerResult{Server: serve}, err
}

// NewAPIServer creates api server by http.Handler from DI container
func NewAPIServer(v *viper.Viper, l *zap.Logger, h http.Handler) (ServerResult, error) {
	return NewHTTPServer(v, "api", h, l)
}

// NewHTTPServer creates http-server that will be embedded into multi-server
func NewHTTPServer(v *viper.Viper, key string, h http.Handler, l *zap.Logger) (ServerResult, error) {
	switch {
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
