package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/response"
)

// ValidationErrorHandler handles validation and binding errors consistently
type ValidationErrorHandler struct{}

// NewValidationErrorHandler creates a new validation error handler
func NewValidationErrorHandler() *ValidationErrorHandler {
	return &ValidationErrorHandler{}
}

// HandleBindingError processes binding and validation errors and sends appropriate responses
func (h *ValidationErrorHandler) HandleBindingError(ctx *gin.Context, err error) {
	var validationErrors validator.ValidationErrors

	switch {
	case errors.As(err, &validationErrors):
		h.sendValidationError(ctx, validationErrors)
	case errors.Is(err, errors.New("EOF")) || err.Error() == "unexpected EOF":
		h.sendBadRequestError(ctx, "Empty or malformed request body")
	default:
		h.sendBadRequestError(ctx, err.Error())
	}
}

// sendValidationError formats and sends validation errors
func (h *ValidationErrorHandler) sendValidationError(ctx *gin.Context, validationErrors validator.ValidationErrors) {
	response := response.BaseErrorResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  h.formatValidationErrors(validationErrors),
	}

	ctx.JSON(http.StatusBadRequest, response)
}

// sendBadRequestError sends a generic bad request error
func (h *ValidationErrorHandler) sendBadRequestError(ctx *gin.Context, message string) {
	response := response.BaseErrorResponse{
		Success: false,
		Message: message,
		Errors:  []response.BaseResponseError{},
	}

	ctx.JSON(http.StatusBadRequest, response)
}

// formatValidationErrors converts validator errors to a structured format
func (h *ValidationErrorHandler) formatValidationErrors(ve validator.ValidationErrors) []response.BaseResponseError {
	errors := make([]response.BaseResponseError, 0, len(ve))

	for _, fieldError := range ve {
		errors = append(errors, response.BaseResponseError{
			Field:   fieldError.Field(),
			Message: h.getValidationErrorMessage(fieldError),
		})
	}

	return errors
}

// getValidationErrorMessage returns user-friendly validation error messages
func (h *ValidationErrorHandler) getValidationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + fe.Param() + " characters long"
	case "max":
		return "Must not exceed " + fe.Param() + " characters"
	case "uuid4":
		return "Must be a valid UUID v4"
	case "ip":
		return "Must be a valid IP address"
	case "ipv4":
		return "Must be a valid IPv4 address"
	case "ipv6":
		return "Must be a valid IPv6 address"
	case "alpha":
		return "Must contain only alphabetic characters"
	case "numeric":
		return "Must contain only numeric characters"
	case "alphanum":
		return "Must contain only alphanumeric characters"
	case "url":
		return "Must be a valid URL"
	case "datetime":
		return "Must be a valid date/time format"
	case "oneof":
		return "Must be one of: " + fe.Param()
	default:
		return "Invalid value for field: " + fe.Tag()
	}
}

// HandleBindError Helper function for backward compatibility
func HandleBindError(ctx *gin.Context, err error) {
	handler := NewValidationErrorHandler()
	handler.HandleBindingError(ctx, err)
}
