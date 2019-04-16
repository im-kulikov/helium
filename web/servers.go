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

	profileResult struct {
		dig.Out

		Handler http.Handler `name:"profile_handler"`
	}

	metricParams struct {
		dig.In

		Handler http.Handler `name:"metric_handler" optional:"true"`
		Viper   *viper.Viper
		Logger  logger.StdLogger
	}

	metricResult struct {
		dig.Out

		Handler http.Handler `name:"metric_handler"`
	}
)

var (
	// ProfileHandlerModule that provides default profile handler
	ProfileHandlerModule = module.Module{
		{Constructor: newProfileHandler},
	}

	// MetricHandlerModule that provides default metric handler
	MetricHandlerModule = module.Module{
		{Constructor: newMetricHandler},
	}

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
	mserv.SetLogger(params.Logger)
	return mserv.New(params.Servers...)
}

func newProfileHandler() profileResult {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return profileResult{Handler: mux}
}

func newProfileServer(p profileParams) ServerResult {
	return newHTTPServer(p.Viper, "pprof", p.Handler, p.Logger)
}

func newMetricHandler() metricResult {
	return metricResult{Handler: promhttp.Handler()}
}

func newMetricServer(p metricParams) ServerResult {
	return newHTTPServer(p.Viper, "metrics", p.Handler, p.Logger)
}

// NewAPIServer creates api server by http.Handler from DI container
func NewAPIServer(v *viper.Viper, l logger.StdLogger, h http.Handler) ServerResult {
	return newHTTPServer(v, "api", h, l)
}

func newHTTPServer(v *viper.Viper, key string, h http.Handler, l logger.StdLogger) ServerResult {
	if h == nil {
		l.Printf("Empty handler for %s server, skip", key)
		return ServerResult{}
	} else if !v.IsSet(key + ".address") {
		l.Printf("Empty bind address for %s server, skip", key)
		return ServerResult{}
	}

	l.Printf("Create %s http server, bind address: %s", key, v.GetString(key+".address"))
	return ServerResult{
		Server: mserv.NewHTTPServer(
			v.GetDuration(key+".shutdown_timeout"),
			&http.Server{
				Addr:              v.GetString(key + ".address"),
				Handler:           h,
				ReadTimeout:       v.GetDuration(key + ".read_timeout"),
				ReadHeaderTimeout: v.GetDuration(key + ".read_header_timeout"),
				WriteTimeout:      v.GetDuration(key + ".write_timeout"),
				IdleTimeout:       v.GetDuration(key + ".idle_timeout"),
				MaxHeaderBytes:    v.GetInt(key + ".max_header_bytes"),
			},
		)}
}
