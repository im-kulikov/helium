package settings

import (
	"net/http"
	"time"

	"github.com/im-kulikov/helium/logger"
	"github.com/spf13/viper"
)

func HTTPServer(key string, router http.Handler) (shutdownTimeout time.Duration, server *http.Server) {
	if !viper.IsSet(key + ".address") {
		logger.G().Warnw("missing address for http server", "server", key)
		return
	}

	shutdownTimeout = viper.GetDuration(key + ".shutdown_timeout")
	server = &http.Server{
		Addr:         viper.GetString(key + ".address"),
		Handler:      router,
		ReadTimeout:  viper.GetDuration(key + ".read_timeout"),
		WriteTimeout: viper.GetDuration(key + ".write_timeout"),
	}
	return
}
