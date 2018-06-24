package main

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/im-kulikov/helium"
	"github.com/im-kulikov/helium/settings"
)

func defaultCommand() helium.Command {
	return helium.Command{
		Name:      "default",
		ShortName: "d",
		Action: func(ctx helium.Context) error {

			spew.Dump(settings.G())

			return nil
		},
	}
}

type app struct{}

func (app) Commands() []helium.Command {
	return []helium.Command{
		// default command
		defaultCommand(),
	}
}

func main() {
	helium.Run(&helium.Config{
		App:       new(app),
		File:      "config.yml", // can be omitted
		Name:      "demo",
		Version:   "1.0.0",
		BuildTime: time.Now().Format(time.RFC3339),
	})
}
