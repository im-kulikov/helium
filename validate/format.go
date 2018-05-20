package validate

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/labstack/echo"
	"gopkg.in/go-playground/validator.v9"
)

type tagParser = func(tag reflect.StructTag) string

var _ = AddTagParsers

var options = []tagParser{
	// Parse json-tag
	func(tag reflect.StructTag) string {
		val := tag.Get("json")
		return strings.Split(val, ",")[0]
	},

	// Parse form-tag
	func(tag reflect.StructTag) string {
		val := tag.Get("form")
		return strings.Split(val, ",")[0]
	},

	// Parse query-tag
	func(tag reflect.StructTag) string {
		val := tag.Get("query")
		return strings.Split(val, ",")[0]
	},

	// Parse form-tag
	func(tag reflect.StructTag) string {
		val := tag.Get("xml")
		return strings.Split(val, ",")[0]
	},

	// Parse yaml-tag
	func(tag reflect.StructTag) string {
		val := tag.Get("yaml")
		return strings.Split(val, ",")[0]
	},

	// Parse param-tag
	func(tag reflect.StructTag) string {
		val := tag.Get("param")
		return strings.Split(val, ",")[0]
	},
}

// AddTagParsers used in fieldName
func AddTagParsers(parser tagParser) {
	options = append(options, parser)
}

// fieldName parse struct for field name in json / form / query tags:
func fieldName(v reflect.Value, field string) string {
	var tp = v.Type()

	if f, ok := tp.FieldByName(field); ok {
		for _, o := range options {
			if val := o(f.Tag); len(val) > 0 && val != "-" {
				return val
			}
		}
	}

	return strings.ToLower(field)
}

// Options to call CheckErrors method
type Options struct {
	Struct    interface{}
	Errors    error
	Formatter func(fields []string) string
}

// defaultFormatter generates "bad `field1`, `field2`"
func defaultFormatter(fields []string) string {
	return "bad `" + strings.Join(fields, "`, `") + "`"
}

// CheckErrors of validator and return formatted errors:
func CheckErrors(opts Options) (ok bool, err error) {
	var fieldsErr validator.ValidationErrors

	if opts.Struct == nil || opts.Errors == nil {
		return
	}

	if opts.Formatter == nil {
		opts.Formatter = defaultFormatter
	}

	if fieldsErr, ok = opts.Errors.(validator.ValidationErrors); ok {
		var (
			fields = make([]string, 0, len(fieldsErr))
			val    = reflect.ValueOf(opts.Struct)
		)

		if val.Kind() == reflect.Ptr && !val.IsNil() {
			val = val.Elem()
		}

		for _, field := range fieldsErr {
			fields = append(fields, fieldName(val, field.Field()))
		}

		err = echo.NewHTTPError(http.StatusBadRequest, opts.Formatter(fields))
	}

	return
}
