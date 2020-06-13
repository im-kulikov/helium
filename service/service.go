package service

import (
	"context"

	"go.uber.org/dig"
	"go.uber.org/zap"
)

type (
	// runnable interface
	runner interface {
		Start(context.Context) error
		Stop() error
	}

	// Service interface
	Service interface {
		runner
		Name() string
	}

	// Group wrapper around group of services
	Group runner

	// Params for service module
	Params struct {
		dig.In

		Logger *zap.Logger
		Group  []Service `group:"services"`
	}

	group struct {
		log   *zap.Logger
		items []Service
	}
)

// create group of services
func newGroup(p Params) Group {
	services := make([]Service, 0, len(p.Group))

	for i := range p.Group {
		if p.Group[i] == nil {
			p.Logger.Warn("ignore nil service", zap.Int("position", i))
			continue
		}

		p.Logger.Info("add service", zap.String("name", p.Group[i].Name()))

		services = append(services, p.Group[i])
	}

	return &group{
		log:   p.Logger,
		items: services,
	}
}

// Start all services
func (s *group) Start(ctx context.Context) error {
	for i, svc := range s.items {
		if svc == nil {
			s.log.Warn("ignore nil service", zap.Int("position", i))

			continue
		}
		s.log.Info("run service", zap.String("name", svc.Name()))

		if err := svc.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Stop all services
func (s *group) Stop() error {
	var lastError error
	for i, svc := range s.items {
		if svc == nil {
			s.log.Warn("ignore nil service",
				zap.Int("position", i))

			continue
		}

		err := svc.Stop()

		s.log.Info("stop service",
			zap.String("name", svc.Name()),
			zap.Error(err))

		if err != nil {
			lastError = err
		}
	}

	return lastError
}
