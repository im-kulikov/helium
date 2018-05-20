package validate

import "github.com/go-playground/validator"

var defaultValidate = &validate{validator.New()}

// validate is wrapper of echo.Validator
type validate struct {
	Validator *validator.Validate
}

// Validator is the interface that wraps the Validate function.
type Validator interface {
	Validate(interface{}) error
}

// G global validator
func G() *validator.Validate {
	return defaultValidate.Validator
}

// Echo validator
func Echo() Validator {
	return defaultValidate
}

// New creates new wrapper of validator for echo.Validator
func New(v *validator.Validate) Validator {
	return &validate{v}
}

// Validate a struct(s) exposed fields, and automatically validates nested struct(s), unless otherwise specified.
func (v *validate) Validate(i interface{}) error {
	return v.Validator.Struct(i)
}

// Echo validator
func (v *validate) Echo() Validator {
	return v
}
