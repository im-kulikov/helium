package settings

import (
	"strings"

	"github.com/spf13/viper"

	"github.com/im-kulikov/helium/module"
)

// Module of config things.
// nolint:gochecknoglobals
var Module = module.Module{
	{Constructor: New},
}

// nolint:gochecknoglobals
var global = viper.New()

// Viper returns global Viper instance.
func Viper() *viper.Viper { return global }

// New init viper settings.
func New(app *Core) (*viper.Viper, error) {
	v := viper.New()
	global = v
	v.SetEnvPrefix(app.Prefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if len(app.File) > 0 {
		v.SetConfigType(app.SafeType())
		v.SetConfigFile(app.File)

		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	return v, nil
}
