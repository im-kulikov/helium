package workers

import (
	"context"

	"github.com/chapsuk/worker"
	"github.com/go-redis/redis"
	"github.com/im-kulikov/helium/module"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

// Module of workers
var Module = module.Module{
	{Constructor: NewWorkersGroup},
	{Constructor: NewWorkers},
}

type (
	// Result returns wrapped workers group for di
	Result struct {
		dig.Out
		Workers []*worker.Worker
	}

	// Params is dependencies for create workers slice
	Params struct {
		dig.In

		Config *viper.Viper
		Logger *zap.Logger
		Redis  *redis.Client `optional:"true"`
		Jobs   map[string]worker.Job
	}

	options struct {
		Viper  *viper.Viper
		Redis  *redis.Client
		Logger *zap.SugaredLogger
		CfgKey string
		Job    worker.Job
	}
)

// NewWorkersGroup returns workers group with injected workers
func NewWorkersGroup(wrks []*worker.Worker) *worker.Group {
	wg := worker.NewGroup()
	wg.Add(wrks...)
	return wg
}

// NewWorkers returns wrapped workers slice builded by config settings
func NewWorkers(p Params) (Result, error) {
	res := Result{}
	for name, job := range p.Jobs {
		wrk, err := workerByConfig(options{
			Viper:  p.Config,
			Redis:  p.Redis,
			Logger: p.Logger.Sugar(),
			CfgKey: name,
			Job:    job,
		})
		if err != nil {
			// all or nothing
			return Result{}, err
		}
		res.Workers = append(res.Workers, wrk)
	}
	return res, nil
}

func workerByConfig(opts options) (*worker.Worker, error) {
	key := "workers." + opts.CfgKey
	if !opts.Viper.IsSet(key) {
		return nil, errors.Wrap(ErrMissingKey, key)
	}

	if opts.Viper.IsSet(key+".disabled") && opts.Viper.GetBool(key+".disabled") {
		return worker.New(func(context.Context) {}), nil
	}

	w := worker.New(opts.Job)

	if opts.Viper.IsSet(key + ".timer") {
		w.ByTimer(opts.Viper.GetDuration(key + ".timer"))
	}
	if opts.Viper.IsSet(key + ".ticker") {
		w.ByTicker(opts.Viper.GetDuration(key + ".ticker"))
	}
	if opts.Viper.IsSet(key + ".cron") {
		w.ByCronSpec(opts.Viper.GetString(key + ".cron"))
	}

	if opts.Viper.IsSet(key + ".lock") {
		if opts.Redis == nil {
			return nil, errors.Wrap(ErrRedisClientNil, opts.CfgKey)

		}
		lockOptions := worker.RedisLockOptions{
			RedisCLI: opts.Redis,
			LockKey:  opts.Viper.GetString(key + ".lock.key"),
			LockTTL:  opts.Viper.GetDuration(key + ".lock.ttl"),
			Logger:   opts.Logger.With("worker", opts.CfgKey),
		}
		if opts.Viper.IsSet(key + ".lock.retry.count") {
			w.WithBsmRedisLock(worker.BsmRedisLockOptions{
				RedisLockOptions: lockOptions,
				RetryCount:       opts.Viper.GetInt(key + ".lock.retry.count"),
				RetryDelay:       opts.Viper.GetDuration(key + ".lock.retry.timeout"),
			})
		} else {
			w.WithRedisLock(lockOptions)
		}
	}

	if opts.Viper.IsSet(key + ".immediately") {
		w.SetImmediately(opts.Viper.GetBool(key + ".immediately"))
	}

	return w, nil
}
