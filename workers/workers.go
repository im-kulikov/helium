package workers

import (
	"context"

	"github.com/chapsuk/worker"
	"github.com/im-kulikov/helium/module"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

type (
	// Params is dependencies for create workers slice
	Params struct {
		dig.In

		Config *viper.Viper
		Jobs   map[string]worker.Job
		Locker worker.Locker `optional:"true"`
	}

	// LockerSettings creates copy of locker and applies settings
	LockerSettings interface {
		Apply(key string, v *viper.Viper) (worker.Locker, error)
	}

	options struct {
		Name   string
		Job    worker.Job
		Viper  *viper.Viper
		Locker worker.Locker
	}
)

// Module of workers
var Module = module.Module{
	{Constructor: NewWorkers},
	{Constructor: NewWorkersGroup},
}

func nopJob(_ context.Context) {}

// NewWorkersGroup returns workers group with injected workers
func NewWorkersGroup(workers []*worker.Worker) *worker.Group {
	var items = make([]*worker.Worker, 0, len(workers))

	for i := range workers {
		if workers[i] != nil {
			items = append(items, workers[i])
		}
	}

	wg := worker.NewGroup()
	wg.Add(items...)
	return wg
}

// NewWorkers returns wrapped workers slice created by config settings
func NewWorkers(p Params) ([]*worker.Worker, error) {
	switch {
	case p.Config == nil:
		return nil, ErrEmptyConfig
	case p.Jobs == nil || len(p.Jobs) == 0:
		return nil, ErrEmptyWorkers
	}

	workers := make([]*worker.Worker, 0, len(p.Jobs))
	for name, job := range p.Jobs {
		wrk, err := workerByConfig(options{
			Viper:  p.Config,
			Locker: p.Locker,
			Name:   name,
			Job:    job,
		})
		if err != nil {
			// all or nothing
			return nil, err
		}
		workers = append(workers, wrk)
	}
	return workers, nil
}

func workerByConfig(opts options) (*worker.Worker, error) {
	key := "workers." + opts.Name

	switch {
	case !opts.Viper.IsSet(key):
		return nil, errors.Wrap(ErrMissingKey, key)
	case opts.Viper.IsSet(key+".disabled") && opts.Viper.GetBool(key+".disabled"):
		return worker.New(nopJob), nil
	case opts.Job == nil:
		return nil, errors.Wrap(ErrEmptyJob, opts.Name)
	}

	w := worker.New(opts.Job)

	if opts.Viper.IsSet(key + ".timer") {
		w = w.ByTimer(opts.Viper.GetDuration(key + ".timer"))
	}
	if opts.Viper.IsSet(key + ".ticker") {
		w = w.ByTicker(opts.Viper.GetDuration(key + ".ticker"))
	}
	if opts.Viper.IsSet(key + ".cron") {
		w = w.ByCronSpec(opts.Viper.GetString(key + ".cron"))
	}
	if opts.Viper.IsSet(key + ".immediately") {
		w = w.SetImmediately(opts.Viper.GetBool(key + ".immediately"))
	}

	if opts.Viper.IsSet(key + ".lock") {
		if opts.Locker == nil {
			return nil, errors.Wrap(ErrEmptyLocker, key)
		} else if l, ok := opts.Locker.(LockerSettings); ok {
			locker, err := l.Apply(key, opts.Viper)
			if err != nil {
				return nil, err
			}

			w = w.WithLock(locker)
		} else {
			w = w.WithLock(opts.Locker)
		}
	}

	return w, nil
}
