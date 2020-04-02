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
		Stop()
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

	services struct {
		log   *zap.Logger
		items []Service
	}
)

// create group of services
func newGroup(p Params) Group {
	return &services{
		log:   p.Logger,
		items: p.Group,
	}
}

// Start all services
func (s *services) Start(ctx context.Context) error {
	for _, svc := range s.items {
		s.log.Info("run service",
			zap.String("name", svc.Name()))

		if err := svc.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Stop all services
func (s *services) Stop() {
	for _, svc := range s.items {
		s.log.Info("stop service",
			zap.String("name", svc.Name()))
		svc.Stop()
	}
}
