package module

import "go.uber.org/dig"

type (
	Module []*Provider

	Provider struct {
		Constructor interface{}
		Options     []dig.ProvideOption
	}
)

// New single module
func New(fn interface{}, opts ...dig.ProvideOption) Module {
	if fn == nil {
		return Module{}
	}

	return Module{
		{
			Constructor: fn,
			Options:     opts,
		},
	}
}

// Append module to target module and return new module
func (m Module) Append(mods ...Module) Module {
	var result = m
	for _, mod := range mods {
		result = append(result, mod...)
	}
	return result
}

// Provide set providers functions to DI container
func Provide(dic *dig.Container, providers Module) error {
	for _, p := range providers {
		if err := dic.Provide(p.Constructor, p.Options...); err != nil {
			return err
		}
	}
	return nil
}
