package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/im-kulikov/helium/cli"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/settings"
)

func defaultCommand() cli.Command {
	return cli.Command{
		Name:      "default",
		ShortName: "d",
		Action: func(ctx *cli.Context) error {

			spew.Dump(settings.ORM())
			spew.Dump(settings.G())

			return nil
		},
	}
}

func commands() []cli.Command {
	return []cli.Command{
		// default command
		defaultCommand(),
	}
}

func main() {
	if err := cli.Run(&cli.Options{
		Name:     "demo",
		Version:  "1.0.0",
		Config:   "config.yml",
		Commands: commands(),
	}); err != nil {
		logger.Panic(err)
	}
}
