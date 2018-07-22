package helium

import (
	"context"

	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	"go.uber.org/dig"
)

type (
	App interface {
		Run(ctx context.Context) error
	}

	Helium struct {
		di *dig.Container
	}
)

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
