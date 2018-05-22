package settings

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var buildTime, buildVersion string

// Init settings
func Init(appName, appVersion, appBuildTime, filename string) error {
	viper.SetConfigFile(filename)
	viper.SetEnvPrefix(appName)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	buildTime, buildVersion = appBuildTime, appVersion

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	setLoggerConfig(appName, appVersion)
	setOrmConfig()
	setRedisConfig()

	return nil
}

// Version of application
func Version() string {
	return buildVersion
}

// BuildTime of application
func BuildTime() string {
	return buildTime
}

// G global config
func G() *viper.Viper {
	return viper.GetViper()
}

func missingKeyErr(key string) error {
	return errors.New("missing config key: " + key)
}
