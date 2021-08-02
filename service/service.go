package service

import (
	"context"
	"time"

	"go.uber.org/dig"
	"go.uber.org/zap"

	"github.com/im-kulikov/helium/group"
)

type (
	runner interface {
		Start(context.Context) error
		Stop(context.Context)
	}

	// Service interface.
	Service interface {
		runner
		Name() string
	}

	// Group wrapper around group of services.
	Group interface {
		Run(context.Context) error
	}

	// Params for service module.
	Params struct {
		dig.In

		Logger   *zap.Logger
		Group    []Service     `group:"services"`
		Shutdown time.Duration `name:"service_shutdown_timeout"`
	}

	multiple struct {
		*zap.Logger
		group.Service
	}
)

// create group of services.
func newGroup(p Params) Group {
	run := &multiple{
		Logger:  p.Logger,
		Service: group.New(group.WithShutdownTimeout(p.Shutdown)),
	}

	p.Logger.Info("added workers", zap.Int("count", len(p.Group)))

	for i := range p.Group {
		if p.Group[i] == nil {
			p.Logger.Warn("ignore nil service", zap.Int("position", i))

			continue
		}

		run.Add(run.prepareActor(p.Group[i]))
	}

	return run
}

func (m *multiple) prepareActor(svc Service) (group.Callback, group.Shutdown) {
	m.Info("add service", zap.String("name", svc.Name()))

	return func(ctx context.Context) error {
			m.Info("run service", zap.String("name", svc.Name()))

			return svc.Start(ctx)
		},

		func(ctx context.Context) {
			m.Info("stop service", zap.String("name", svc.Name()))

			svc.Stop(ctx)
		}
}
