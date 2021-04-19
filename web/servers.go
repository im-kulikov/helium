package web

import (
	"net"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/im-kulikov/helium/internal"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
)

type (
	// APIParams struct
	APIParams struct {
		dig.In

		Config   *viper.Viper
		Logger   *zap.Logger
		Handler  http.Handler `optional:"true"`
		Listener net.Listener `name:"api_listener" optional:"true"`
	}

	HTTPParams struct {
		dig.In

		Config   *viper.Viper
		Logger   *zap.Logger
		Name     string       `name:"http_name" optional:"true"`
		Key      string       `name:"http_config" optional:"true"`
		Handler  http.Handler `name:"http_handler" optional:"true"`
		Listener net.Listener `name:"http_listener" optional:"true"`
	}

	// ServerResult struct
	ServerResult struct {
		dig.Out

		Server service.Service `group:"services"`
	}

	profileParams struct {
		dig.In

		Logger   *zap.Logger
		Viper    *viper.Viper
		Handler  http.Handler `name:"pprof_handler" optional:"true"`
		Listener net.Listener `name:"pprof_listener" optional:"true"`
	}

	metricParams struct {
		dig.In

		Logger   *zap.Logger
		Viper    *viper.Viper
		Handler  http.Handler `name:"metric_handler" optional:"true"`
		Listener net.Listener `name:"metric_listener" optional:"true"`
	}

	grpcParams struct {
		dig.In

		Logger   *zap.Logger
		Viper    *viper.Viper
		Name     string       `name:"grpc_name" optional:"true"`
		Key      string       `name:"grpc_config" optional:"true"`
		Server   *grpc.Server `name:"grpc_server" optional:"true"`
		Listener net.Listener `name:"grpc_listener" optional:"true"`
	}
)

const (
	apiServer     = "api"
	gRPCServer    = "grpc"
	profileServer = "pprof"
	metricsServer = "metric"
)

var (
	// DefaultServersModule of web base structs.
	DefaultServersModule = module.Combine(
		DefaultGRPCModule,
		ProfilerModule,
		MetricsModule,
		APIModule,
	)

	// APIModule defines API server module.
	APIModule = module.New(NewAPIServer)

	// ProfilerModule defines pprof server module.
	ProfilerModule = module.New(newProfileServer)

	// MetricsModule defines prometheus server module.
	MetricsModule = module.New(newMetricServer)

	// DefaultGRPCModule defines default gRPC server module.
	DefaultGRPCModule = module.New(newDefaultGRPCServer)

	// ErrEmptyLogger is raised when empty logger passed into New function.
	ErrEmptyLogger = internal.Error("empty logger")
)

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
	return NewHTTPServer(HTTPParams{
		Config:   p.Viper,
		Logger:   p.Logger,
		Name:     profileServer,
		Key:      profileServer,
		Handler:  mux,
		Listener: p.Listener,
	})
}

func newMetricServer(p metricParams) (ServerResult, error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	if p.Handler != nil {
		mux.Handle("/", p.Handler)
	}
	return NewHTTPServer(HTTPParams{
		Config:   p.Viper,
		Logger:   p.Logger,
		Name:     metricsServer,
		Key:      metricsServer,
		Handler:  mux,
		Listener: p.Listener,
	})
}

// NewAPIServer creates api server by http.Handler from DI container
func NewAPIServer(p APIParams) (ServerResult, error) {
	return NewHTTPServer(HTTPParams{
		Config:   p.Config,
		Logger:   p.Logger,
		Name:     apiServer,
		Key:      apiServer,
		Handler:  p.Handler,
		Listener: p.Listener,
	})
}

func newDefaultGRPCServer(p grpcParams) (ServerResult, error) {
	if p.Key == "" {
		p.Key = gRPCServer
	}

	if p.Name == "" {
		p.Name = "default_grpc"
	}

	switch {
	case p.Logger == nil:
		return ServerResult{}, ErrEmptyLogger
	case p.Viper == nil:
		p.Logger.Info("Empty config for gRPC server, skip")
		return ServerResult{}, nil
	case p.Viper.GetBool(p.Key + ".disabled"):
		p.Logger.Info("Server disabled",
			zap.String("name", p.Name))
		return ServerResult{}, nil
	case p.Server == nil:
		p.Logger.Info("Empty server, skip",
			zap.String("name", p.Name))
		return ServerResult{}, nil
	}

	var address string
	options := []GRPCOption{
		GRPCName(p.Name),
		GRPCWithLogger(p.Logger),
		GRPCListener(p.Listener),
	}

	if p.Viper.GetBool(p.Key + ".skip_errors") {
		options = append(options, GRPCSkipErrors())
	}

	if p.Viper.IsSet(p.Key + ".network") {
		options = append(options, GRPCListenNetwork(p.Viper.GetString(p.Key+".network")))
	}

	if p.Viper.IsSet(p.Key + ".address") {
		address = p.Viper.GetString(p.Key + ".address")
		options = append(options, GRPCListenAddress(address))
	}

	if p.Listener != nil {
		address = p.Listener.Addr().String()
	}

	serve, err := NewGRPCService(p.Server, options...)

	p.Logger.Info("Creates gRPC server", zap.String("name", p.Key), zap.String("address", address))

	return ServerResult{Server: serve}, err
}

// NewHTTPServer creates http-server that will be embedded into multi-server
func NewHTTPServer(p HTTPParams) (ServerResult, error) {
	switch {
	case p.Logger == nil:
		return ServerResult{}, ErrEmptyLogger
	case p.Key == "" || p.Config == nil:
		p.Logger.Info("Empty config or key for http server, skip", zap.String("key", p.Key))
		return ServerResult{}, nil
	case p.Handler == nil:
		p.Logger.Info("Empty handler, skip", zap.String("name", p.Key))
		return ServerResult{}, nil
	case p.Config.GetBool(p.Key + ".disabled"):
		p.Logger.Info("Server disabled", zap.String("name", p.Key))
		return ServerResult{}, nil
	}

	options := []HTTPOption{
		HTTPName(p.Key),
		HTTPListener(p.Listener),
		HTTPWithLogger(p.Logger),
	}

	var address string
	if p.Config.IsSet(p.Key + ".address") {
		address = p.Config.GetString(p.Key + ".address")
		options = append(options, HTTPListenAddress(address))
	}

	if p.Config.IsSet(p.Key + ".network") {
		options = append(options, HTTPListenNetwork(p.Config.GetString(p.Key+".network")))
	}

	if p.Config.IsSet(p.Key + ".skip_errors") {
		options = append(options, HTTPSkipErrors())
	}

	hServer := &http.Server{Handler: p.Handler}
	if p.Config.IsSet(p.Key + ".read_timeout") {
		hServer.ReadTimeout = p.Config.GetDuration(p.Key + ".read_timeout")
	}

	if p.Config.IsSet(p.Key + ".read_header_timeout") {
		hServer.ReadHeaderTimeout = p.Config.GetDuration(p.Key + ".read_header_timeout")
	}

	if p.Config.IsSet(p.Key + ".write_timeout") {
		hServer.WriteTimeout = p.Config.GetDuration(p.Key + ".write_timeout")
	}

	if p.Config.IsSet(p.Key + ".idle_timeout") {
		hServer.IdleTimeout = p.Config.GetDuration(p.Key + ".idle_timeout")
	}

	if p.Config.IsSet(p.Key + ".max_header_bytes") {
		hServer.MaxHeaderBytes = p.Config.GetInt(p.Key + ".max_header_bytes")
	}

	if p.Listener != nil {
		address = p.Listener.Addr().String()
	}

	serve, err := NewHTTPService(hServer, options...)
	if err != nil {
		return ServerResult{}, err
	}

	p.Logger.Info("creates http server", zap.String("name", p.Name), zap.String("address", address))

	return ServerResult{Server: serve}, nil
}
