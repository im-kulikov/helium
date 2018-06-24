package helium

import (
	"context"
	"fmt"
	"os"

	"github.com/chapsuk/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/settings"
	"github.com/im-kulikov/helium/validate"
	"github.com/im-kulikov/helium/web"
	"github.com/pkg/errors"
	"go.uber.org/dig"
	"gopkg.in/urfave/cli.v1"
)

type (
	App interface {
		Commands() []Command
	}

	InvokeOption  = dig.InvokeOption
	ProvideOption = dig.ProvideOption

	Container interface {
		Invoke(function interface{}, opts ...InvokeOption) error
		Provide(constructor interface{}, opts ...ProvideOption) error
	}

	ActionFunc func(ctx Context) error

	Command struct {
		// Init method to invoke dependencies
		Init func(di Container) error
		// An action to execute after any subcommands are run, but after the subcommand has finished
		// It is run even if Action() panics
		After cli.AfterFunc
		// An action to execute before any sub-subcommands are run, but after the context is ready
		// If a non-nil error is returned, no sub-subcommands are run
		Before cli.BeforeFunc
		// The function to call when this command is invoked
		Action ActionFunc
		// The name of the command
		Name string
		// short name of the command. Typically one character (deprecated, use `Aliases`)
		ShortName string
		// A list of aliases for the command
		Aliases []string
		// A short description of the usage of this command
		Usage string
	}

	ctx struct {
		command  *cli.Context
		graceful context.Context
	}

	Context interface {
		Command() *cli.Context
		Graceful() context.Context
	}

	Config struct {
		App       App
		File      string
		Name      string
		Version   string
		BuildTime string
	}

	helium struct {
		app App
		cli *cli.App
		cfg *Config
		di  *dig.Container
	}
)

var defaultModules = []interface{}{
	settings.Redis,
	settings.ORM,
	validate.New,
	web.NewBinder,
	web.NewLogger,
	web.NewEngine,
	web.NewServers,
}

func (c *ctx) Command() *cli.Context {
	return c.command
}

func (c *ctx) Graceful() context.Context {
	return c.graceful
}

func Run(cfg *Config) {
	instance := &helium{
		app: cfg.App,
		cli: cli.NewApp(),
		di:  dig.New(),
		cfg: cfg,
	}

	instance.mustDefaults()

	instance.cli.Name = cfg.Name
	instance.cli.Version = cfg.Version
	instance.cli.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Usage:       "path to config file",
			Value:       cfg.File,
			Destination: &cfg.File,
		},
	}

	instance.cli.Before = func(ctx *cli.Context) error {
		if err := settings.Init(cfg.Name, cfg.Version, cfg.BuildTime, cfg.File); err != nil {
			instance.panic(errors.Wrap(err, "can't initialize settings"))
		}

		if err := logger.Init(settings.Logger()); err != nil {
			instance.panic(errors.Wrap(err, "can't initialize logger"))
		}

		for _, module := range defaultModules {
			if err := instance.di.Provide(module); err != nil {
				instance.panic(dig.RootCause(err))
			}
		}

		return nil
	}

	items := instance.app.Commands()
	commands := make([]cli.Command, 0, len(items))

	for _, item := range items {
		// closure..
		func(command Command) {
			commands = append(commands, cli.Command{
				After: command.After,
				Before: func(c *cli.Context) error {
					if command.Init != nil {
						if err := command.Init(instance.di); err != nil {
							instance.panic(dig.RootCause(err))
						}
					}

					if command.Before != nil {
						return command.Before(c)
					}

					return nil
				},
				Name:      command.Name,
				ShortName: command.ShortName,
				Aliases:   command.Aliases,
				Usage:     command.Usage,
				Action: func(c *cli.Context) error {
					if err := command.Action(&ctx{
						command:  c,
						graceful: grace.ShutdownContext(context.Background()),
					}); err != nil {
						instance.panic(err)
					}
					return nil
				},
			})
		}(item)
	}

	instance.cli.Commands = commands

	if err := instance.cli.Run(os.Args); err != nil {
		instance.panic(err)
	}
}

func (h *helium) mustDefaults() {
	if len(h.cfg.Name) == 0 {
		h.cfg.Name = "unknown"
	}

	if len(h.cfg.BuildTime) == 0 {
		h.cfg.BuildTime = "unknown"
	}

	if len(h.cfg.Version) == 0 {
		h.cfg.Version = "unknown"
	}

	if len(h.cfg.File) == 0 {
		h.cfg.File = "config.yml"
	}
	if h.cfg == nil {
		h.panic("Config can't be nil")
	}

	if h.cfg.App == nil {
		h.panic("Config.App can't be nil")
	}
}

func (h *helium) panic(err interface{}) {
	if logger.G() != nil {
		logger.G().Fatalw("can't run application", "error", err)
	} else {
		fmt.Printf(`{"app_name": "%s", "app_version": "%s", "msg": "start app error", "error": "%s"}`,
			h.cfg.Name,
			h.cfg.Version,
			err,
		)
		os.Exit(1)
	}
}
