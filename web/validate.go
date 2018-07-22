package web

import (
	"gopkg.in/go-playground/validator.v9"
)

type (
	Func       = validator.Func
	FieldLevel = validator.FieldLevel
	Validator  interface {
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

func NewValidator() Validator {
	return WrapValidator(validator.New())
}

func WrapValidator(v *validator.Validate) Validator {
	return &validate{
		v: v,
	}
}
