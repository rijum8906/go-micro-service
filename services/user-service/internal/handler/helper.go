package handler

import (
	"github.com/go-playground/validator/v10"
)

func formatValidationErrors(ve validator.ValidationErrors) []BaseResponseError {
	errors := make([]BaseResponseError, 0, len(ve))

	for _, fe := range ve {
		errors = append(errors, BaseResponseError{
			Field:   fe.Field(),
			Message: validationErrorMessage(fe),
		})
	}

	return errors
}

// validationErrorMessage converts validator tags into human-friendly messages.
func validationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email address"
	case "min":
		return "Must be at least " + fe.Param() + " characters"
	case "max":
		return "Must be at most " + fe.Param() + " characters"
	case "uuid4":
		return "Invalid UUID format"
	case "ip":
		return "Invalid IP address"
	case "alpha":
		return "Must contain only alphabetic characters"
	default:
		return "Invalid value"
	}
}
