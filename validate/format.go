package validate

import (
	"fmt"
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

	// Parse xml-tag
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

type (
	// Options to call CheckErrors method
	Options struct {
		Struct    interface{}
		Errors    error
		Formatter func(fields []*FieldError) string
	}

	// FieldError contains field name and validator error
	FieldError struct {
		Field     string
		Message   string
		Validator validator.FieldError
	}
)

func (f FieldError) Error() string {
	return f.Validator.(error).Error()
}

// defaultFormatter generates "bad `field1`, `field2`"
func defaultFormatter(fields []*FieldError) string {
	items := make([]string, 0, len(fields))
	withMessage := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field.Message) > 0 {
			withMessage = append(withMessage, fmt.Sprintf("`%s` %s",
				field.Field,
				field.Message,
			))
			continue
		}

		items = append(items, field.Field)
	}

	var result []string

	if len(items) > 0 {
		result = append(result, "bad value of `"+strings.Join(items, "`, `")+"`")
	}

	if len(withMessage) > 0 {
		result = append(result, strings.Join(withMessage, "; "))
	}

	return strings.Join(result, "; ")
}

func messageParse(v reflect.Value, field string) string {
	var tp = v.Type()

	// Parse message-tag
	msgParse := func(tag reflect.StructTag) string {
		val := tag.Get("message")
		return strings.Split(val, ",")[0]
	}

	if f, ok := tp.FieldByName(field); ok {
		if val := msgParse(f.Tag); len(val) > 0 && val != "-" {
			return val
		}
	}

	return ""
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
			fields = make([]*FieldError, 0, len(fieldsErr))
			val    = reflect.ValueOf(opts.Struct)
		)

		if val.Kind() == reflect.Ptr && !val.IsNil() {
			val = val.Elem()
		}

		for _, field := range fieldsErr {
			fields = append(fields, &FieldError{
				Field:     fieldName(val, field.Field()),
				Message:   messageParse(val, field.Field()),
				Validator: field,
			})
		}

		err = echo.NewHTTPError(http.StatusBadRequest, opts.Formatter(fields))
	}

	return
}
