package helium

import (
	"context"
	stdlog "log"

	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

type (
	// App implementation for helium
	App interface {
		Run(ctx context.Context) error
	}

	// Helium struct
	Helium struct {
		di *dig.Container
	}
)

// New helium instance
func New(cfg *settings.App, mod module.Module) (*Helium, error) {
	h := &Helium{
		di: dig.New(),
	}

	if cfg != nil {
		mod = append(mod, cfg.Provider())
	}

	if err := module.Provide(h.di, mod); err != nil {
		return nil, err
	}

	return h, nil
}

// Run trying invoke app instance from DI container and start app with Run call
func (p Helium) Run() error {
	return p.di.Invoke(func(ctx context.Context, app App) error {
		return app.Run(ctx)
	})
}

// Catch errors
func Catch(err error) {
	if err == nil {
		return
	}

	v := viper.New()
	log, logErr := logger.
		NewLogger(logger.NewLoggerConfig(v), &settings.App{
			Name:         "",
			BuildVersion: "",
		})
	if logErr != nil {
		stdlog.Fatal(err)
	} else {
		log.
			Sugar().
			Fatalw("Can't run app",
				"error", err)
	}
}
