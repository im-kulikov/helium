package cli

import (
	"os"

	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/settings"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type (
	// App alias
	App = cli.App

	// Flag alias
	Flag = cli.Flag

	// StringFlag alias
	StringFlag = cli.StringFlag

	// IntFlag alias
	IntFlag = cli.IntFlag

	// BoolFlag alias
	BoolFlag = cli.BoolFlag

	// Command alias
	Command = cli.Command

	// Context alias
	Context = cli.Context

	// Options for cli-application
	Options struct {
		Name      string
		BuildTime string // time in RFC format
		Version   string
		Config    string // default config path
		Flags     []Flag
		Commands  []Command
	}
)

var (
	// ErrEmptyConfig when given empty options
	ErrEmptyConfig = errors.New("cli empty config")
)

func getDefaults(opts *Options) {
	if len(opts.Config) == 0 {
		opts.Config = "config.yaml"
	}

	if len(opts.Flags) == 0 {
		opts.Flags = make([]Flag, 0, 1)
	}

	opts.Flags = append(opts.Flags, cli.StringFlag{
		Name:        "config",
		Usage:       "path to config file",
		Value:       opts.Config,
		Destination: &opts.Config,
	})
}

// Run cli-application
func Run(opts *Options) error {
	if opts == nil {
		return ErrEmptyConfig
	}

	getDefaults(opts)

	app := cli.NewApp()
	app.Flags = opts.Flags
	app.Commands = opts.Commands

	app.Before = func(ctx *cli.Context) error {
		if err := settings.Init(opts.Name, opts.Version, opts.BuildTime, opts.Config); err != nil {
			logger.Panic(err)
		}

		if err := logger.Init(settings.Logger(), nil); err != nil {
			logger.Panic(err)
		}

		return nil
	}

	return app.Run(os.Args)
}
