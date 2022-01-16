package group

import (
	"context"
	"errors"
	"sync"
	"time"
)

type (
	group struct {
		ignore   []error
		services []service
		shutdown time.Duration
	}

	service struct {
		callback Callback
		shutdown Shutdown
	}

	// Shutdown function that receives shutdown context and allows to gracefully stop an service.
	Shutdown func(context.Context)

	// Callback function that will be called on service starts.
	Callback func(context.Context) error

	// Service collects services and runs them concurrently.
	// - when any service returns, all services will be stopped.
	// - when context canceled or deadlined all services will be stopped.
	Service interface {
		Add(Callback, Shutdown) Service
		Run(context.Context) error
	}
)

const defaultShutdown = time.Second * 5

var (
	_ Service = (*group)(nil)

	// nolint:gochecknoglobals
	defaultIgnoredErrors = []error{
		context.Canceled,
		context.DeadlineExceeded,
	}
)

// New creates and configures Service by passed Option's.
func New(options ...Option) Service {
	runner := &group{
		shutdown: defaultShutdown,
		ignore:   defaultIgnoredErrors,
	}

	for _, o := range options {
		o(runner)
	}

	return runner
}

// Add an service (callback and shutdown) to the group.
// Canceling context shutdowns all running services.
// The first service (callback function) to return shutdowns all running services.
// The context.Context passed into shutdown function needed to gracefully shutdown services.
func (g *group) Add(callback Callback, stopper Shutdown) Service {
	g.services = append(g.services, service{
		callback: callback,
		shutdown: stopper,
	})

	return g
}

func (g *group) checkAndIgnore(err error) error {
	for i := range g.ignore {
		if errors.Is(err, g.ignore[i]) {
			return nil
		}
	}

	return err
}

// Run allows to run all services (callback function).
// - method blocks until all services will be stopped.
// - when context will be canceled or deadline exceeded we calls shutdown for services.
// - when the first service (callback function) returns, all other services will be notified to stop.
func (g *group) Run(ctx context.Context) error {
	if len(g.services) == 0 {
		return nil
	}

	var (
		cnt int
		err error
		res = make(chan error, len(g.services))
	)

	// we should add cancel to prevent service freeze
	top, cancel := context.WithCancel(ctx)

	// run all services
	for i := range g.services {
		go func(callback Callback) { res <- callback(top) }(g.services[i].callback)
	}

	// wait for context.Done() or error will be received:
	select {
	case err = <-res:
		cnt = 1 // first error received, ignore it in future
	case <-top.Done():
		err = top.Err()
	}

	cancel()

	// prepare graceful context to stop
	grace, stop := context.WithTimeout(context.Background(), g.shutdown)
	defer stop()

	// we should wait until all services will gracefully stopped
	wg := new(sync.WaitGroup)
	defer wg.Wait()

	wg.Add(len(g.services))
	// notify all services to stop
	for i := range g.services {
		go func(shutdown Shutdown) {
			defer wg.Done()

			shutdown(grace)
		}(g.services[i].shutdown)
	}

	// wait when all services will stop
	for i := cnt; i < cap(res); i++ {
		<-res
	}

	// return only first error except ignored errors
	return g.checkAndIgnore(err)
}
