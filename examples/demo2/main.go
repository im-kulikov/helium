package main

import (
	"os"

	"github.com/im-kulikov/helium"
	"github.com/im-kulikov/helium/examples/demo2/app"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	"github.com/urfave/cli"
)

const (
	name      = "demo2"
	config    = "config.yml"
	version   = "1.0.0"
	buildTime = "now"
)

func run(mod module.Module) cli.ActionFunc {
	return func(*cli.Context) error {
		h, err := helium.New(&settings.App{
			File:         config,
			Name:         name,
			BuildTime:    version,
			BuildVersion: buildTime,
		}, mod)

		if err != nil {
			return err
		}

		return h.Run()
	}
}

func main() {
	c := cli.NewApp()
	c.Name = name
	c.Version = version
	c.Commands = cli.Commands{
		{
			Name:      "serve",
			ShortName: "s",
			Action:    run(app.ServeCommandModule),
		},
		{
			Name:      "test",
			ShortName: "t",
			Action:    run(app.TestCommandModule),
		},
	}

	if err := c.Run(os.Args); err != nil {
		panic(err)
	}
}
