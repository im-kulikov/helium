package web

import (
	"context"
	"os"

	"github.com/im-kulikov/helium/internal"
	"github.com/im-kulikov/helium/service"
	"go.uber.org/zap"
)

type (
	runner struct {
		services []service.Service
		logger   *zap.Logger
	}
)

const (
	// ErrEmptyLogger is raised when empty logger passed into New function.
	ErrEmptyLogger = internal.Error("empty logger")

	// ErrEmptyServices is raised when empty services passed into New function.
	ErrEmptyServices = internal.Error("empty services")
)

var fatal = os.Exit

// New gets logger and services to create multiple service runner.
func New(log *zap.Logger, services ...service.Service) (service.Service, error) {
	if log == nil {
		return nil, ErrEmptyLogger
	}

	multi := &runner{
		logger:   log,
		services: make([]service.Service, 0, len(services)),
	}

	for i := range services {
		if services[i] == nil {
			continue
		}

		multi.services = append(multi.services, services[i])
	}

	if len(multi.services) == 0 {
		return nil, ErrEmptyServices
	}

	return multi, nil
}

// Name returns name of the service
func (m *runner) Name() string { return "multi-runner" }

// Start tries to start every server and returns error
// if something went wrong.
func (m *runner) Start(ctx context.Context) error {
	for i := range m.services {
		if err := m.services[i].Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Stop tries to stop services, logs every error,
// and returns last error.
func (m *runner) Stop() error {
	var lastErr error
	for i := range m.services {
		if err := m.services[i].Stop(); err != nil {
			lastErr = err

			m.logger.Error("could not stop server",
				zap.Int("index", i),
				zap.Error(err))
		}
	}

	return lastErr
}
