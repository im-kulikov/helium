package web

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/im-kulikov/helium/module"
	"github.com/labstack/echo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestViper() *viper.Viper {
	v := viper.New()
	v.SetDefault("api.debug", true)
	return v
}

func jsonUnmarshalError() error {
	return &json.UnmarshalTypeError{
		Value:  "",
		Type:   reflect.TypeOf("something"),
		Offset: 0,
		Struct: "",
		Field:  "",
	}
}

func jsonSyntaxError() error {
	return &json.SyntaxError{
		Offset: 50,
	}
}

func xmlUnsuportTypeError() error {
	return &xml.UnsupportedTypeError{
		Type: reflect.TypeOf("something"),
	}
}

func xmlSyntaxError() error {
	return &xml.SyntaxError{}
}

type myTestCustomError struct {
	Message string
	Fail    bool
}

func (e myTestCustomError) Error() string {
	return e.Message
}
func (e myTestCustomError) FormatResponse(ctx echo.Context) error {
	if e.Fail {
		return errors.New(e.Message)
	}
	return ctx.String(http.StatusBadRequest, e.Message)
}

func customError() error {
	return myTestCustomError{Message: "this is custom error"}
}

func customErrorFail() error {
	return myTestCustomError{
		Message: "this is custom error",
		Fail:    true,
	}
}

func httpError() error {
	return echo.NewHTTPError(http.StatusBadRequest, "some bad request")
}

func httpInternalError() error {
	return echo.NewHTTPError(http.StatusInternalServerError, "some internal error")
}

func unknownError() error {
	return errors.New("unknown error")
}

type testBuffer struct {
	*bytes.Buffer
}

func (testBuffer) Sync() error {
	return nil
}

func newTestLogger(rw zapcore.WriteSyncer) *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), rw, zap.DebugLevel)
	return zap.New(core)
}

func TestEngine(t *testing.T) {
	Convey("Engine", t, func() {
		Convey("try create and check new engine", func() {
			var (
				v = NewValidator()
				b = NewBinder(v)
				z = zap.L()
				l = NewLogger(z)
				c = newTestViper()

				_ = c

				err error
				dic = dig.New()
			)

			Convey("create engine with empty params", func() {
				err = module.Provide(dic, module.Module{
					{Constructor: NewEngine},
				})

				So(err, ShouldBeNil)

				err = dic.Invoke(func(e *echo.Echo) {
					So(e.Binder, ShouldNotEqual, b)
					So(e.Logger, ShouldNotEqual, l)
					So(e.Validator, ShouldNotEqual, v)
					So(e.Debug, ShouldBeFalse)
				})

				So(err, ShouldBeNil)
			})

			Convey("create engine with binder, zap.logger, validate and debug", func() {
				err = module.Provide(dic, module.Module{
					{Constructor: func() echo.Validator { return v }},
					{Constructor: func() echo.Binder { return b }},
					{Constructor: func() *zap.Logger { return z }},
					{Constructor: func() *viper.Viper { return c }},
					{Constructor: NewEngine},
				})

				So(err, ShouldBeNil)

				err = dic.Invoke(func(e *echo.Echo) {
					So(e.Binder, ShouldEqual, b)
					_, ok := e.Logger.(*echoLogger)
					So(ok, ShouldBeTrue)
					So(e.Validator, ShouldEqual, v)
					So(e.Debug, ShouldBeTrue)
				})

				So(err, ShouldBeNil)
			})

			Convey("create engine with binder, logger, validate and debug", func() {
				err = module.Provide(dic, module.Module{
					{Constructor: func() echo.Validator { return v }},
					{Constructor: func() echo.Binder { return b }},
					{Constructor: func() echo.Logger { return l }},
					{Constructor: func() *viper.Viper { return c }},
					{Constructor: NewEngine},
				})

				So(err, ShouldBeNil)

				err = dic.Invoke(func(e *echo.Echo) {
					So(e.Binder, ShouldEqual, b)
					So(e.Logger, ShouldEqual, l)
					So(e.Validator, ShouldEqual, v)
					So(e.Debug, ShouldBeTrue)
				})

				So(err, ShouldBeNil)
			})
		})

		Convey("try to capture errors", func() {
			var (
				buf      = new(bytes.Buffer)
				z        = newTestLogger(testBuffer{Buffer: buf})
				e        = NewEngine(EngineParams{Logger: z})
				req, err = http.NewRequest("POST", "/some-url", nil)
				rec      = httptest.NewRecorder()
				ctx      = e.NewContext(req, rec)
			)

			Convey("try to capture json.Unmarshal errors", func() {

				So(err, ShouldBeNil)
				err = jsonUnmarshalError()
				So(err, ShouldBeError)
				captureError(z.Sugar())(err, ctx)
				So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				So(rec.Body.String(), ShouldContainSubstring, "JSON parse error: expected=")
			})

			Convey("try to capture json.Syntax errors", func() {

				So(err, ShouldBeNil)
				err = jsonSyntaxError()
				So(err, ShouldBeError)
				captureError(z.Sugar())(err, ctx)
				So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				So(rec.Body.String(), ShouldContainSubstring, "JSON parse error: offset=")
			})

			Convey("try to capture xml.Unmarshal errors", func() {

				So(err, ShouldBeNil)
				err = xmlUnsuportTypeError()
				So(err, ShouldBeError)
				captureError(z.Sugar())(err, ctx)
				So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				So(rec.Body.String(), ShouldContainSubstring, "XML parse error: type=")
			})

			Convey("try to capture xml.Syntax errors", func() {

				So(err, ShouldBeNil)
				err = xmlSyntaxError()
				So(err, ShouldBeError)
				captureError(z.Sugar())(err, ctx)
				So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				So(rec.Body.String(), ShouldContainSubstring, "XML parse error: line=")
			})

			Convey("try to capture custom errors", func() {

				So(err, ShouldBeNil)
				err = customError()
				So(err, ShouldBeError)
				captureError(z.Sugar())(err, ctx)
				So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				So(rec.Body.String(), ShouldContainSubstring, "this is custom error")
			})

			Convey("try to capture custom errors and fail", func() {

				So(err, ShouldBeNil)
				err = customErrorFail()
				So(err, ShouldBeError)
				captureError(z.Sugar())(err, ctx)
				So(rec.Body.Len(), ShouldEqual, 0)
				So(buf.String(), ShouldContainSubstring, "Capture error")
				So(buf.String(), ShouldContainSubstring, "this is custom error")
			})

			Convey("try to capture http.Error", func() {

				So(err, ShouldBeNil)
				err = httpError()
				So(err, ShouldBeError)
				captureError(z.Sugar())(err, ctx)
				So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				So(rec.Body.String(), ShouldContainSubstring, "some bad request")
			})

			Convey("try to capture http.Error 500", func() {

				So(err, ShouldBeNil)
				err = httpInternalError()
				So(err, ShouldBeError)
				captureError(z.Sugar())(err, ctx)
				So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				So(rec.Body.String(), ShouldContainSubstring, "some internal error")
				So(buf.String(), ShouldContainSubstring, "Request error")
			})

			Convey("try to capture unknown Error", func() {

				So(err, ShouldBeNil)
				err = unknownError()
				So(err, ShouldBeError)
				captureError(z.Sugar())(err, ctx)
				So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				So(rec.Body.String(), ShouldContainSubstring, http.StatusText(http.StatusBadRequest))
			})
		})
	})
}
