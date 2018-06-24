package settings

import (
	"github.com/im-kulikov/helium/redis"
	"github.com/spf13/viper"
)

var redisConfig = &redis.Config{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
	PoolSize: 2,
}

func fetchRedisConfig(key string, config *redis.Config) {
	if config == nil {
		return
	}

	if !viper.IsSet(key) {
		return
	}

	if viper.IsSet(key + ".address") {
		redisConfig.Addr = viper.GetString(key + ".address")
	}

	if viper.IsSet(key + ".password") {
		redisConfig.Password = viper.GetString(key + ".password")
	}

	if viper.IsSet(key + ".database") {
		redisConfig.DB = viper.GetInt(key + ".database")
	}

	if viper.IsSet(key + ".pool_size") {
		redisConfig.PoolSize = viper.GetInt(key + ".pool_size")
	}
}

func setRedisConfig() {
	fetchRedisConfig("redis", redisConfig)
}

// Redis connection
func Redis() (*redis.Client, error) {
	return redis.New(redisConfig)
}
