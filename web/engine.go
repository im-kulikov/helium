package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/validate"
	"github.com/labstack/echo"
)

type (
	errorResponse struct {
		Error  string   `json:"error"`
		Stack  []string `json:"stack"`
		Result []string `json:"result"`
	}

	CustomError interface {
		FormatResponse(ctx echo.Context)
	}
)

func captureError(err error, ctx echo.Context) {
	var (
		message string
		code    = http.StatusBadRequest
		result  = make([]string, 0)
		trace   = make([]string, 0)
	)

	switch custom := err.(type) {
	case CustomError:
		custom.FormatResponse(ctx)
		return
	case *json.UnmarshalTypeError:
		message = fmt.Sprintf("JSON parse error: expected=%v, got=%v, offset=%v", custom.Type, custom.Value, custom.Offset)
	case *json.SyntaxError:
		message = fmt.Sprintf("JSON parse error: offset=%v, error=%v", custom.Offset, custom.Error())
	case *xml.UnsupportedTypeError:
		message = fmt.Sprintf("XML parse error: type=%v, error=%v", custom.Type, custom.Error())
	case *xml.SyntaxError:
		message = fmt.Sprintf("XML parse error: line=%v, error=%v", custom.Line, custom.Error())
	case *echo.HTTPError:
		code = custom.Code
		message = custom.Message.(string)
	default:
		message = http.StatusText(code)
	}

	// Capture errors:
	if code >= http.StatusInternalServerError {
		logger.G().Errorw("Request error", "error", err)
	}

	if errJSON := ctx.JSON(code, errorResponse{
		Error:  message,
		Stack:  trace,
		Result: result,
	}); errJSON != nil {
		logger.G().Errorw("Capture error", "error", err)
	}
}

func NewEngine(l echo.Logger, b echo.Binder, v validate.Validator) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = false
	e.Logger = l
	e.Validator = v
	e.Binder = b
	e.HTTPErrorHandler = captureError

	return e
}
