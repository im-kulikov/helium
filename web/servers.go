package web

import (
	"net/http"

	"github.com/chapsuk/mserv"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

type (
	APIParams struct {
		dig.In

		Config  *viper.Viper
		Logger  logger.StdLogger
		Handler http.Handler `optional:"true"`
	}

	MultiServerParams struct {
		dig.In

		Logger  logger.StdLogger
		Servers []mserv.Server `group:"web_server"`
	}

	ServerResult struct {
		dig.Out

		Server mserv.Server `group:"web_server"`
	}
)

var (
	// ServersModule of web base structs
	ServersModule = module.Module{
		{Constructor: NewAPIServer},
		{Constructor: NewMetricsServer},
		{Constructor: NewPprofServer},
		{Constructor: NewMultiServer},
	}
)

// NewMultiServer returns new multi servers group
func NewMultiServer(params MultiServerParams) mserv.Server {
	mserv.SetLogger(params.Logger)
	return mserv.New(params.Servers...)
}

// NewPprofServer returns wrapped pprof http server
func NewPprofServer(v *viper.Viper, l logger.StdLogger) ServerResult {
	return newHTTPServer(v, "pprof", http.DefaultServeMux, l)
}

// NewMetricsServer returns wrapped prometheus http server
func NewMetricsServer(v *viper.Viper, l logger.StdLogger) ServerResult {
	return newHTTPServer(v, "metrics", promhttp.Handler(), l)
}

// NewAPIServerParams params for create api server by http.Handler from DI container
func NewAPIServer(v *viper.Viper, l logger.StdLogger, h http.Handler) ServerResult {
	return newHTTPServer(v, "api", h, l)
}

func newHTTPServer(v *viper.Viper, key string, h http.Handler, l logger.StdLogger) ServerResult {
	if !v.IsSet(key + ".address") {
		l.Printf("Empty bind address for %s server, skip", key)
		return ServerResult{}
	}
	if h == nil {
		l.Printf("Empty handler for %s server, skip", key)
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
