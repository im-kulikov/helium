package workers

import (
	"context"

	"github.com/chapsuk/worker"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type (
	// Params is dependencies for create workers slice
	Params struct {
		dig.In

		Config *viper.Viper
		Logger *zap.Logger
		Jobs   map[string]worker.Job
		Locker worker.Locker `optional:"true"`
	}

	// Out params for DI
	Out struct {
		dig.Out

		Workers []service.Service `group:"services,flatten"`
	}

	wrapper struct {
		*worker.Worker

		name string
		done chan struct{}
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

var (
	_ = Module // prevent unused

	// Module of workers
	Module = module.New(NewWorkers)
)

func nopJob(_ context.Context) {}

// Start wrapped worker
func (w *wrapper) Start(ctx context.Context) error {
	go func() {
		w.Worker.Run(ctx)
		close(w.done)
	}()

	return nil
}

// Stop wrapped worker
func (w *wrapper) Stop() error {
	<-w.done

	return nil
}

// Name of the wrapped worker
func (w *wrapper) Name() string {
	return w.name
}

// NewWorkers returns wrapped workers slice created by config settings
func NewWorkers(p Params) (Out, error) {
	var result Out

	switch {
	case p.Config == nil:
		return result, ErrEmptyConfig
	case p.Jobs == nil || len(p.Jobs) == 0:
		return result, ErrEmptyWorkers
	}

	result.Workers = make([]service.Service, 0, len(p.Jobs))
	for name, job := range p.Jobs {
		wrk, err := workerByConfig(options{
			Viper:  p.Config,
			Locker: p.Locker,
			Job:    job,
			Name:   name,
		})
		if err != nil {
			// all or nothing
			return result, err
		}

		p.Logger.Info("Create new worker", zap.String("name", name))

		result.Workers = append(result.Workers, &wrapper{
			Worker: wrk,
			name:   "workers." + name,
			done:   make(chan struct{}),
		})
	}

	return result, nil
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
