package settings

import (
	"github.com/im-kulikov/helium/orm"
	"github.com/spf13/viper"
)

var ormConfig = &orm.Config{
	Addr:     "localhost:5432",
	User:     "postgres",
	Password: "postgres",
	Database: "postgres",
	PoolSize: 2,
}

func setOrmConfig() {
	if !viper.IsSet("database") {
		return
	}

	if viper.IsSet("database.address") {
		ormConfig.Addr = viper.GetString("database.address")
	}

	if viper.IsSet("database.user") {
		ormConfig.User = viper.GetString("database.user")
	}

	if viper.IsSet("database.password") {
		ormConfig.Password = viper.GetString("database.password")
	}

	if viper.IsSet("database.database") {
		ormConfig.Database = viper.GetString("database.database")
	}

	if viper.IsSet("database.debug") {
		ormConfig.Debug = viper.GetBool("database.debug")
	}

	if viper.IsSet("database.pool_size") {
		ormConfig.PoolSize = viper.GetInt("database.pool_size")
	}
}

// ORM config
func ORM() *orm.Config {
	return ormConfig
}
