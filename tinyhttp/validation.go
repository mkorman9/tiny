package tinyhttp

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"reflect"
	"strings"
)

// DefaultValidator is the default instance of validator.Validate.
var DefaultValidator = validator.New()

// ValidationError denotes an error in payload validation.
type ValidationError struct {
	// Field is a name of the field that contains an error.
	Field string `json:"field"`

	// Tag is a name of the tag that trigger an error.
	Tag string `json:"tag"`
}

// BindBody tries to parse provided request body and validate resulting object using the DefaultValidator.
func BindBody(c *fiber.Ctx, out any) []ValidationError {
	if err := c.BodyParser(out); err != nil {
		return []ValidationError{
			{Field: "body", Tag: "json"},
		}
	}

	if err := DefaultValidator.Struct(out); err != nil {
		return ExtractValidatorErrors(err)
	}

	return nil
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
	DefaultValidator.RegisterTagNameFunc(func(field reflect.StructField) string {
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
