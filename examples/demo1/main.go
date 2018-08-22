package main

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/im-kulikov/helium"
	"github.com/im-kulikov/helium/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

var mod = module.New(newApp).
	Append(
		settings.Module,
		logger.Module,
		grace.Module)

type App struct {
	v *viper.Viper
}

func newApp(v *viper.Viper) helium.App {
	return &App{v: v}
}

func (a App) Run(ctx context.Context) error {
	spew.Dump(a.v.AllSettings())

	return nil
}

func main() {
	h, err := helium.New(&settings.App{
		File:         "config.yml",
		Type:         "yml",
		Name:         "demo",
		BuildTime:    "now",
		BuildVersion: "dev",
	}, mod)

	if err != nil {
		panic(dig.RootCause(err))
	}

	if err := h.Run(); err != nil {
		panic(err)
	}
}
