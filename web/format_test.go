package web

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	. "github.com/smartystreets/goconvey/convey"
)

type test1 struct {
	A int `json:"a_custom" validate:"gt=0" message:"must be greater than 0"`
	B int `form:"b_custom" validate:"required" message:"is required"`
	C int
	D int `query:"someValue" validate:"required"`
	E int `json:"-" form:"-" query:"-" param:"-" xml:"-" yaml:"-" validate:"required"`
	F int `json:"-" param:"f_custom" xml:"-" yaml:"-" validate:"required"`
	G int `xml:"g_custom" yaml:"-" validate:"required"`
	H int `yaml:"h_custom" validate:"required"`
}

type testCase struct {
	Name   string
	Struct interface{}
	Error  error
}

var testCases = []testCase{
	{
		Name:   "validate errors for all fields",
		Struct: test1{A: -1},
		Error:  echo.NewHTTPError(http.StatusBadRequest, "bad value of `someValue`, `e`, `f_custom`, `g_custom`, `h_custom`; `a_custom` must be greater than 0; `b_custom` is required"),
	},

	{
		Name:   "validate errors for A field",
		Struct: test1{A: 0, B: 1, D: 1, E: 1, F: 1, G: 1, H: 1},
		Error:  echo.NewHTTPError(http.StatusBadRequest, "`a_custom` must be greater than 0"),
	},

	{
		Name:   "validate errors for B field",
		Struct: test1{A: 1, D: 1, E: 1, F: 1, G: 1, H: 1},
		Error:  echo.NewHTTPError(http.StatusBadRequest, "`b_custom` is required"),
	},

	{
		Name:   "validate errors for D, E, F, G, H fields",
		Struct: test1{A: 1, B: 1},
		Error:  echo.NewHTTPError(http.StatusBadRequest, "bad value of `someValue`, `e`, `f_custom`, `g_custom`, `h_custom`"),
	},

	{
		Name:   "validate errors must be empty",
		Struct: test1{A: 1, B: 1, D: 1, E: 1, F: 1, G: 1, H: 1},
		Error:  nil,
	},
}

func TestCheckErrors(t *testing.T) {
	Convey("Prepare validator", t, func() {
		v := NewValidator()

		So(v, ShouldNotBeNil)

		So(len(testCases) > 0, ShouldBeTrue)
		for _, test := range testCases {
			Convey(test.Name, func() {
				errValidate := v.Validate(test.Struct)

				if test.Error == nil {
					So(errValidate, ShouldBeNil)
				} else {
					So(errValidate, ShouldBeError)
				}

				ok, err := CheckErrors(ValidateParams{
					Struct: test.Struct,
					Errors: errValidate,
				})

				if test.Error == nil {
					So(ok, ShouldBeFalse)
					So(err, ShouldBeNil)
				} else {
					So(ok, ShouldBeTrue)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, test.Error.Error())
				}
			})
		}
	})
}
