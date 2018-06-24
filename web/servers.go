package web

import (
	"net/http"

	"github.com/chapsuk/mserv"
	"github.com/im-kulikov/helium/settings"
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewServers(handler http.Handler, log echo.Logger) mserv.Server {
	mserv.SetLogger(log)
	return mserv.New(
		mserv.NewHTTPServer(settings.HTTPServer("pprof", http.DefaultServeMux)),
		mserv.NewHTTPServer(settings.HTTPServer("metrics", promhttp.Handler())),
		mserv.NewHTTPServer(settings.HTTPServer("api", handler)),
	)
}
