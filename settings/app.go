package settings

import (
	"go.uber.org/dig"

	"github.com/im-kulikov/helium/module"
)

type (
	// Defaults is callback that allows to setup application before run.
	// Helium calls `defaults` handler at the end in `New` method.
	Defaults interface{}

	// Core configuration.
	Core struct {
		File         string
		Type         string
		Name         string
		Prefix       string
		BuildTime    string
		BuildVersion string
	}
)

// DIProvider wrap di into provider.
func DIProvider(di *dig.Container) *module.Provider {
	return &module.Provider{
		Constructor: func() *dig.Container { return di },
	}
}

// Provider - wrap app config to provider.
func (a *Core) Provider() *module.Provider {
	return &module.Provider{
		Constructor: func() *Core { return a },
	}
}

// SafeType returns config type, default config type: yaml.
// returns yml if config type not supported.
// nolint:goconst
func (a Core) SafeType() string {
	switch a.Type {
	case "toml", "yml", "yaml":
		return a.Type
	default:
		return "yaml"
	}
}
