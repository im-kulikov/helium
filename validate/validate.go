package validate

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

func New() Validator {
	return &validate{
		v: validator.New(),
	}
}
