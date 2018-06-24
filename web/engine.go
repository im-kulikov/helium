package web

import (
	"github.com/im-kulikov/helium/validate"
	"github.com/labstack/echo"
)

func NewEngine(l echo.Logger, b echo.Binder, v validate.Validator) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = false
	e.Logger = l
	e.Validator = v
	e.Binder = b

	return e
}
