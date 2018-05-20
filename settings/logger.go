package settings

import (
	"github.com/im-kulikov/helium/logger"
	"github.com/spf13/viper"
)

var loggerConfig = &logger.Config{
	Level:  "debug",
	Format: "console",
}

func setLoggerConfig(appName, appVersion string) {
	loggerConfig.AppName = appName
	loggerConfig.AppVersion = appVersion

	if viper.IsSet("logger.level") {
		loggerConfig.Level = viper.GetString("logger.level")
	}

	if viper.IsSet("logger.format") {
		format := viper.GetString("logger.format")
		switch format {
		case "console":
			fallthrough
		case "json":
			loggerConfig.Format = format
		}
	}
}

// Logger config
func Logger() *logger.Config {
	return loggerConfig
}
