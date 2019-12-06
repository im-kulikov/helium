package web

import (
	"github.com/im-kulikov/helium/internal"
	"go.uber.org/zap"
)

type (
	// Service interface that allows start and stop
	Service interface {
		Start() error
		Stop() error
	}

	runner struct {
		services []Service
		logger   *zap.Logger
	}
)

const (
	// ErrEmptyLogger is raised when empty logger passed into New function.
	ErrEmptyLogger = internal.Error("empty logger")

	// ErrEmptyServices is raised when empty services passed into New function.
	ErrEmptyServices = internal.Error("empty services")
)

// New gets logger and services to create multiple service runner.
func New(log *zap.Logger, services ...Service) (Service, error) {
	if log == nil {
		return nil, ErrEmptyLogger
	}

	multi := &runner{
		logger:   log,
		services: make([]Service, 0, len(services)),
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

// Start tries to start every server and returns error
// if something went wrong.
func (m *runner) Start() error {
	for i := range m.services {
		if err := m.services[i].Start(); err != nil {
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
