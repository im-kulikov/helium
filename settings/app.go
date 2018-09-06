package settings

import "github.com/im-kulikov/helium/module"

// Core configuration
type Core struct {
	File         string
	Type         string
	Name         string
	Prefix       string
	BuildTime    string
	BuildVersion string
}

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
