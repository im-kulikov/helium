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

func setRedisConfig() {
	if !viper.IsSet("redis") {
		return
	}

	if viper.IsSet("redis.address") {
		redisConfig.Addr = viper.GetString("redis.address")
	}

	if viper.IsSet("redis.password") {
		redisConfig.Password = viper.GetString("redis.password")
	}

	if viper.IsSet("redis.database") {
		redisConfig.DB = viper.GetInt("redis.database")
	}

	if viper.IsSet("redis.pool_size") {
		ormConfig.PoolSize = viper.GetInt("database.pool_size")
	}
}

// Redis config
func Redis() *redis.Config {
	return redisConfig
}
