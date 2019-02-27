package settings

import (
	"strings"

	"github.com/im-kulikov/helium/module"
	"github.com/spf13/viper"
)

// Module of config things
var Module = module.Module{
	{Constructor: New},
}

// New init viper settings
func New(app *Core) (*viper.Viper, error) {
	v := viper.New()
	if len(app.Prefix) > 0 {
		v.SetEnvPrefix(app.Prefix)
	} else {
		v.SetEnvPrefix(app.Name)
	}
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if len(app.File) > 0 {
		v.SetConfigType(app.SafeType())
		v.SetConfigFile(app.File)
		err := v.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}
