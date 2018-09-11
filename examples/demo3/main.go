package main

import (
	"strings"

	"github.com/im-kulikov/helium"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func main() {
	h, err := helium.New(&helium.Settings{
		Name:         "demo3",
		Prefix:       "DM3",
		File:         "config.yml",
		BuildVersion: "dev",
	}, module.Module{}.Append(
		settings.Module,
		logger.Module,
	))
	err = dig.RootCause(err)
	helium.Catch(err)
	err = h.Invoke(runner)
	err = dig.RootCause(err)
	helium.Catch(err)
}

func runner(v *viper.Viper, l *zap.Logger) {
	l.Info("app :: test command")

	tree("", v.AllSettings(), l)
}

func tree(key string, v interface{}, l *zap.Logger) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, v := range val {
			var keys []string

			if key != "" {
				keys = append(keys, key)
			}

			keys = append(keys, k)
			tree(strings.Join(keys, "."), v, l)
		}
	default:
		l.Info("item",
			zap.Any(key, val))
	}
}
