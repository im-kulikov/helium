package settings

import (
	"github.com/im-kulikov/helium/module"
	"go.uber.org/dig"
)

type (
	// Defaults is callback that allows to setup application before run.
	// Helium calls `defaults` handler at the end in `New` method.
	Defaults func(di *dig.Container) error

	// Core configuration
	Core struct {
		File         string
		Type         string
		Name         string
		Prefix       string
		BuildTime    string
		BuildVersion string
	}
)

// Provider - wrap app config to provider
func (a *Core) Provider() *module.Provider {
	return &module.Provider{
		Constructor: func() *Core { return a },
	}
}

// SafeType returns config type, default config type: yml
// returns yml if config type not supported
func (a Core) SafeType() string {
	switch t := a.Type; t {
	case "toml":
	case "yml", "yaml":
	default:
		return "yml"
	}
	return a.Type
}
