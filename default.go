package helium

import (
	"context"

	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
	"go.uber.org/zap"
)

type defaultApp struct {
	log *zap.Logger
	svc service.Group
}

var (
	// DefaultApp defines default helium application
	DefaultApp = module.New(newDefaultApp)

	_     = DefaultApp
	_ App = (*defaultApp)(nil)
)

func newDefaultApp(log *zap.Logger, svc service.Group) App {
	return defaultApp{
		log: log,
		svc: svc,
	}
}

// Run an application
func (d defaultApp) Run(ctx context.Context) error {
	d.log.Info("starting services")

	if err := d.svc.Start(ctx); err != nil {
		return err
	}

	d.log.Info("app successfully run")
	<-ctx.Done()

	d.log.Info("stopping services")
	if err := d.svc.Stop(); err != nil {
		return err
	}

	d.log.Info("application stopped")

	return nil
}
