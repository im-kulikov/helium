package settings

import (
	"github.com/chapsuk/worker"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/redis"
	"github.com/spf13/viper"
)

// Worker returns workers instance by config settings
// looking to `workers.<name>` config key
func Worker(name string, job worker.Job) (*worker.Worker, error) {
	key := "workers." + name
	if !viper.IsSet(key) {
		return nil, missingKeyErr(key)
	}

	w := worker.New(job)

	if viper.IsSet(key + ".timer") {
		w.ByTimer(viper.GetDuration(key + ".timer"))
	}

	if viper.IsSet(key + ".ticker") {
		w.ByTicker(viper.GetDuration(key + ".ticker"))
	}

	if viper.IsSet(key + ".cron") {
		w.ByCronSpec(viper.GetString(key + ".cron"))
	}

	if viper.IsSet(key + ".lock") {

		if viper.IsSet(key + ".lock.redis") {
			var (
				rdKey  = viper.GetString(key + ".lock.redis")
				rdConf *redis.Config
				cli    *redis.Client
				err    error
			)

			if rdKey == "redis" {
				cli, err = Redis()
			} else {
				rdConf = new(redis.Config)
				fetchRedisConfig(rdKey, rdConf)
				cli, err = redis.New(rdConf)
			}

			if err != nil {
				logger.G().Errorw("connect to worker redis lock error", "error", err)
				if viper.GetBool(key + ".critical") {
					return nil, err
				}
			}

			ropts := worker.RedisLockOptions{
				RedisCLI: cli,
				LockKey:  viper.GetString(key + ".lock.key"),
				LockTTL:  viper.GetDuration(key + ".lock.ttl"),
				Logger:   logger.G().With("worker", name),
			}

			if viper.IsSet(key + ".lock.retry") {
				w.WithBsmRedisLock(worker.BsmRedisLockOptions{
					RedisLockOptions: ropts,
					RetryCount:       viper.GetInt(key + ".lock.retry.count"),
					RetryDelay:       viper.GetDuration(key + ".lock.retry.timeout"),
				})
			} else {
				w.WithRedisLock(ropts)
			}
		}
	}

	if viper.IsSet(key + ".immediately") {
		w.SetImmediately(viper.GetBool(key + ".immediately"))
	}

	return w, nil
}
