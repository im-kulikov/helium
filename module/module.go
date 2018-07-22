package module

import "go.uber.org/dig"

type (
	Module []*Provider

	Provider struct {
		Constructor interface{}
		Options     []dig.ProvideOption
	}
)

// Append module to target module and return new module
func (m Module) Append(mod Module) Module {
	return append(m, mod...)
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
