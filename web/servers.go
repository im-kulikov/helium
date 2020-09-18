package web

import (
	"net/http"
	"net/http/pprof"

	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
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
		Servers []service.Service `group:"services"`
	}

	// ServerResult struct
	ServerResult struct {
		dig.Out

		Server service.Service `group:"services"`
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
		Name   string       `name:"grpc_name" optional:"true"`
		Key    string       `name:"grpc_config" optional:"true"`
		Server *grpc.Server `name:"grpc_server" optional:"true"`
	}
)

var (
	// DefaultServersModule of web base structs.
	DefaultServersModule = module.Combine(
		DefaultGRPCModule,
		ProfilerModule,
		MetricsModule,
		APIModule,
		MultiServeModule,
	)

	// APIModule defines API server module.
	APIModule = module.New(NewAPIServer)

	// MultiServeModule defines multi serve module.
	MultiServeModule = module.New(NewMultiServer)

	// ProfilerModule defines pprof server module.
	ProfilerModule = module.New(newProfileServer)

	// MetricsModule defines prometheus server module.
	MetricsModule = module.New(newMetricServer)

	// DefaultGRPCModule defines default gRPC server module.
	DefaultGRPCModule = module.New(newDefaultGRPCServer)
)

// NewMultiServer returns new multi servers group
func NewMultiServer(p MultiServerParams) (service.Service, error) { return New(p.Logger, p.Servers...) }

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

// NewAPIServer creates api server by http.Handler from DI container
func NewAPIServer(p APIParams) (ServerResult, error) {
	return NewHTTPServer(p.Config, "api", p.Handler, p.Logger)
}

func newDefaultGRPCServer(p grpcParams) (ServerResult, error) {
	if p.Name == "" {
		p.Name = "default"
	}

	switch {
	case p.Logger == nil:
		return ServerResult{}, ErrEmptyLogger
	case p.Key == "":
		p.Logger.Info("Empty config key for gRPC server, skip")
		return ServerResult{}, nil
	case p.Viper == nil:
		p.Logger.Info("Empty config for gRPC server, skip")
		return ServerResult{}, nil
	case p.Server == nil:
		p.Logger.Info("Empty server, skip",
			zap.String("name", p.Name))
		return ServerResult{}, nil
	case p.Viper.IsSet(p.Key + ".disabled"):
		p.Logger.Info("Disabled, skip",
			zap.String("name", p.Name))
		return ServerResult{}, nil
	}

	options := []GRPCOption{
		GRPCName(p.Name),
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

	p.Logger.Info("Creates gRPC server",
		zap.String("name", p.Key),
		zap.String("address", p.Viper.GetString(p.Key+".address")))

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

	options = append(options, HTTPName(key))

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
