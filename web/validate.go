package web

import (
	"gopkg.in/go-playground/validator.v9"
)

type (
	// Func accepts a FieldLevel interface for all validation needs. The return
	// value should be true when validation succeeds.
	Func = validator.Func

	// FieldLevel contains all the information and helper functions
	// to validate a field
	FieldLevel = validator.FieldLevel

	// Validator to implement custom echo.Validator
	Validator interface {
		Validate(i interface{}) error
		Register(tag string, fn Func) error
	}

	validate struct {
		v *validator.Validate
	}
)

func (v *validate) Validate(i interface{}) error {
	return v.v.Struct(i)
}

func (v *validate) Register(tag string, fn Func) error {
	return v.v.RegisterValidation(tag, fn)
}

// NewValidator returns custom echo.Validator
func NewValidator() Validator {
	return WrapValidator(validator.New())
}

// WrapValidator wraps v9.validator
func WrapValidator(v *validator.Validate) Validator {
	return &validate{
		v: v,
	}
}
