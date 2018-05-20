package settings

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Init settings
func Init(appName, appVersion, filename string) error {
	viper.SetConfigFile(filename)
	viper.SetEnvPrefix(appName)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	setLoggerConfig(appName, appVersion)
	setOrmConfig()
	setRedisConfig()

	return nil
}

// G global config
func G() *viper.Viper {
	return viper.GetViper()
}

func missingKeyErr(key string) error {
	return errors.New("missing config key: " + key)
}
