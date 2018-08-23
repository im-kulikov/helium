package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type (
	errorResponse struct {
		Error  string   `json:"error"`
		Stack  []string `json:"stack"`
		Result []string `json:"result"`
	}

	// EngineParams struct
	EngineParams struct {
		dig.In

		Config     *viper.Viper   `optional:"true"`
		Binder     echo.Binder    `optional:"true"`
		Logger     *zap.Logger    `optional:"true"`
		EchoLogger echo.Logger    `optional:"true"`
		Validator  echo.Validator `optional:"true"`
	}

	// CustomError interface
	CustomError interface {
		FormatResponse(ctx echo.Context) error
	}
)

func captureError(log *zap.SugaredLogger) echo.HTTPErrorHandler {
	return func(err error, ctx echo.Context) {
		var (
			message string
			code    = http.StatusBadRequest
			result  = make([]string, 0)
			trace   = make([]string, 0)
		)

		switch custom := err.(type) {
		case CustomError:
			if err = custom.FormatResponse(ctx); err != nil {
				log.Errorw("Capture error", "error", err)
			}
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
			log.Errorw("Request error", "error", err)
		}

		if errJSON := ctx.JSON(code, errorResponse{
			Error:  message,
			Stack:  trace,
			Result: result,
		}); errJSON != nil {
			log.Errorw("Capture error",
				"errorJson", errJSON,
				"error", err)
		}
	}
}

// NewEngine returns configured echo engine
func NewEngine(params EngineParams) *echo.Echo {
	e := echo.New()

	e.Debug = false
	e.HidePort = true
	e.HideBanner = true

	if params.Config != nil && params.Config.GetBool("api.debug") {
		e.Debug = true
	}

	if params.Binder != nil {
		e.Binder = params.Binder
	}

	if params.EchoLogger != nil {
		e.Logger = params.EchoLogger
	}

	if params.Logger != nil {
		e.Logger = NewLogger(params.Logger)
		e.HTTPErrorHandler = captureError(params.Logger.Sugar())
	}

	if params.Validator != nil {
		e.Validator = params.Validator
	}

	return e
}
