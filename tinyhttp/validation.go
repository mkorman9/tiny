package tinyhttp

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

// ValidationError denotes an error in payload validation.
type ValidationError struct {
	Field string
	Tag   string
}

// ExtractValidatorErrors tries to extract an array of ValidationError from given error.
func ExtractValidatorErrors(err error) []ValidationError {
	if v, ok := err.(validator.ValidationErrors); ok {
		var result []ValidationError

		for _, e := range v {
			fieldName := extractFieldName(e)
			result = append(result, ValidationError{Field: fieldName, Tag: e.Tag()})
		}

		return result
	}

	return nil
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			fieldName := resolveTag(field, "json")

			if fieldName == "" {
				fieldName = resolveTag(field, "form")
			}

			if fieldName == "" {
				fieldName = resolveTag(field, "header")
			}

			if fieldName == "" {
				fieldName = resolveTag(field, "uri")
			}

			if fieldName == "" {
				fieldName = field.Name
			}

			return fieldName
		})
	}
}

func extractFieldName(fieldError validator.FieldError) string {
	return strings.Join(strings.Split(fieldError.Namespace(), ".")[1:], ".")
}

func resolveTag(field reflect.StructField, tag string) string {
	tagValue := field.Tag.Get(tag)
	if tagValue == "" {
		return ""
	}

	name := strings.SplitN(tagValue, ",", 2)[0]

	if name == "-" {
		return ""
	}

	return name
}
