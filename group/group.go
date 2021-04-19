package group

import (
	"context"
	"errors"
	"time"
)

type (
	group struct {
		actors []actor
		ignore []error
		period time.Duration
	}

	actor struct {
		callback Callback
		shutdown Shutdown
	}

	// Shutdown function that receives shutdown context and allows to gracefully stop an actor.
	Shutdown func(context.Context)

	// Callback function that will be called on actor starts.
	Callback func(context.Context) error

	// Service collects an actors and runs them concurrently.
	// - when any actor returns, all actors will be stopped.
	// - when context canceled or deadlined all actors will be stopped.
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
		period: defaultShutdown,
		ignore: defaultIgnoredErrors,
	}

	for _, o := range options {
		o(runner)
	}

	return runner
}

// Add an actors (callback and shutdown) to the group.
// Canceling context shutdowns all running actors.
// The first actor (function) to return shutdowns all running actors.
// The context.Context passed into shutdown function needed to gracefully shutdown
// web-server or something else.
func (g *group) Add(runner Callback, stopper Shutdown) Service {
	g.actors = append(g.actors, actor{
		callback: runner,
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

// Run allows to run all actors (callback).
// - method blocks until all actors will be stopped.
// - when context will be canceled or deadline exceeded we calls shutdown for actors.
// - when the first actor (callback) returns, all other actors will be notified to stop.
func (g *group) Run(ctx context.Context) error {
	if len(g.actors) == 0 {
		return nil
	}

	errs := make(chan error, len(g.actors))

	// run all actors
	for i := range g.actors {
		go func(item *actor) {
			errs <- item.callback(ctx)
		}(&g.actors[i])
	}

	var (
		cnt int
		err error
		top context.Context
	)

	// wait for context.Done() or error will be received:
	select {
	case err = <-errs:
		cnt = 1 // first error received, ignore it in future
		top = ctx
	case <-ctx.Done():
		err = ctx.Err()
		top = context.Background()
	}

	// prepare graceful context to stop
	grace, stop := context.WithTimeout(top, g.period)
	defer stop()

	// notify all actors to stop
	for i := range g.actors {
		g.actors[i].shutdown(grace)
	}

	// wait when all actors will stop
	for i := cnt; i < cap(errs); i++ {
		<-errs
	}

	// return only first error except ignored errors
	return g.checkAndIgnore(err)
}
