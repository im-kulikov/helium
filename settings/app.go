package settings

import "github.com/im-kulikov/helium/module"

// App configuration
type App struct {
	File         string
	Type         string
	Name         string
	BuildTime    string
	BuildVersion string
}

// Provider - wrap app config to provider
func (a *App) Provider() *module.Provider {
	return &module.Provider{
		Constructor: func() *App { return a },
	}
}

// SafeType returns config type, default config type: yml
// returns yml if config type not supported
func (a App) SafeType() string {
	switch t := a.Type; t {
	case "toml":
	case "yml", "yaml":
	default:
		return "yml"
	}
	return a.Type
}
