package main

import (
	"context"
	"net/http"

	"github.com/chapsuk/mserv"
	"github.com/im-kulikov/helium"
	"github.com/im-kulikov/helium/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	"github.com/im-kulikov/helium/web"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func main() {
	h, err := helium.New(&helium.Settings{
		Name:         "demo3",
		Prefix:       "DM3",
		File:         "config.yml",
		BuildVersion: "dev",
	}, module.Module{
		{Constructor: handler},
	}.Append(
		grace.Module,
		settings.Module,
		web.ServersModule,
		logger.Module,
	))
	err = dig.RootCause(err)
	helium.Catch(err)
	err = h.Invoke(runner)
	err = dig.RootCause(err)
	helium.Catch(err)
}

func handler() http.Handler {
	h := http.NewServeMux()
	h.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return h
}

func runner(s mserv.Server, l *zap.Logger, ctx context.Context) {
	l.Info("run servers")
	s.Start()

	l.Info("application started")
	<-ctx.Done()

	l.Info("stop servers")
	s.Stop()

	l.Info("application stopped")
}
