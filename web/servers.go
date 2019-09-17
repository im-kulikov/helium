package web

import (
	"net/http"
	"net/http/pprof"

	"github.com/chapsuk/mserv"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

type (
	// APIParams struct
	APIParams struct {
		dig.In

		Config  *viper.Viper
		Logger  logger.StdLogger
		Handler http.Handler `optional:"true"`
	}

	// MultiServerParams struct
	MultiServerParams struct {
		dig.In

		Logger  logger.StdLogger
		Servers []mserv.Server `group:"web_server"`
	}

	// ServerResult struct
	ServerResult struct {
		dig.Out

		Server mserv.Server `group:"web_server"`
	}

	profileParams struct {
		dig.In

		Handler http.Handler `name:"profile_handler" optional:"true"`
		Viper   *viper.Viper
		Logger  logger.StdLogger
	}

	metricParams struct {
		dig.In

		Handler http.Handler `name:"metric_handler" optional:"true"`
		Viper   *viper.Viper
		Logger  logger.StdLogger
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
func NewMultiServer(params MultiServerParams) mserv.Server {
	return mserv.New(params.Servers...)
}

func newProfileServer(p profileParams) ServerResult {
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

func newMetricServer(p metricParams) ServerResult {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	if p.Handler != nil {
		mux.Handle("/", p.Handler)
	}
	return NewHTTPServer(p.Viper, "metrics", mux, p.Logger)
}

// NewAPIServer creates api server by http.Handler from DI container
func NewAPIServer(v *viper.Viper, l logger.StdLogger, h http.Handler) ServerResult {
	return NewHTTPServer(v, "api", h, l)
}

// NewHTTPServer creates http-server that will be embedded into multi-server
func NewHTTPServer(v *viper.Viper, key string, h http.Handler, l logger.StdLogger) ServerResult {
	switch {
	case h == nil:
		l.Printf("Empty handler for %s server, skip", key)
		return ServerResult{}
	case v.GetBool(key + ".disabled"):
		l.Printf("Server %s disabled", key)
		return ServerResult{}
	case !v.IsSet(key + ".address"):
		l.Printf("Empty bind address for %s server, skip", key)
		return ServerResult{}
	}

	l.Printf("Create %s http server, bind address: %s", key, v.GetString(key+".address"))
	return ServerResult{
		Server: mserv.NewHTTPServer(
			&http.Server{
				Addr:              v.GetString(key + ".address"),
				Handler:           h,
				ReadTimeout:       v.GetDuration(key + ".read_timeout"),
				ReadHeaderTimeout: v.GetDuration(key + ".read_header_timeout"),
				WriteTimeout:      v.GetDuration(key + ".write_timeout"),
				IdleTimeout:       v.GetDuration(key + ".idle_timeout"),
				MaxHeaderBytes:    v.GetInt(key + ".max_header_bytes"),
			},
			mserv.HTTPShutdownTimeout(v.GetDuration(key+".shutdown_timeout")),
		)}
}
