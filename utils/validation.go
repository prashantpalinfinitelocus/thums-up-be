package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func FormatValidationErrors(err error) map[string][]string {
	fieldErrors := make(map[string][]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			field := getJSONFieldName(fieldError)
			message := getErrorMessage(fieldError)
			fieldErrors[field] = append(fieldErrors[field], message)
		}
	}

	return fieldErrors
}

func getJSONFieldName(fe validator.FieldError) string {
	return toSnakeCase(fe.Field())
}

func getErrorMessage(fe validator.FieldError) string {
	field := getJSONFieldName(fe)

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, fe.Param())
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if i > 1 && str[i-1] >= 'A' && str[i-1] <= 'Z' {
				result.WriteRune(r + ('a' - 'A'))
			} else {
				result.WriteRune('_')
				result.WriteRune(r + ('a' - 'A'))
			}
		} else {
			if r >= 'A' && r <= 'Z' {
				result.WriteRune(r + ('a' - 'A'))
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidPhoneNumber(phone string) bool {
	matched, _ := regexp.MatchString(`^[0-9]{10}$`, phone)
	return matched
}

