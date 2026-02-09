package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	appError "github.com/rijum8906/go-micro-service/packages/common/errors"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services"
)

func formatValidationErrors(ve validator.ValidationErrors) []dto.BaseResponseError {
	errors := make([]dto.BaseResponseError, 0, len(ve))

	for _, fe := range ve {
		errors = append(errors, dto.BaseResponseError{
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

// handleBindError differentiates between validation errors and malformed JSON.
func handleBindError(ctx *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		ctx.JSON(http.StatusBadRequest, dto.BaseErrorResponse{
			Success: false,
			Errors:  formatValidationErrors(ve),
			Message: "Invalid request body",
		})
		return
	}

	// Covers invalid JSON, wrong types, etc.
	ctx.JSON(http.StatusBadRequest, dto.BaseErrorResponse{
		Success: false,
		Message: err.Error(),
	})
}

func parseAppError(err *appError.AppError) gin.H {
	return gin.H{
		"success": false,
		"message": err.Message,
		"errors":  err.Errors,
	}
}

func extractReqMetadata(deviceID string, ctx *gin.Context) dto.RequestMetadata {
	return dto.RequestMetadata{
		DeviceID:  deviceID,
		UserAgent: ctx.Request.UserAgent(),
		IPAddr:    ctx.ClientIP(),
	}
}

func extractAuthzMeatadata(ctx *gin.Context) (dto.AuthzMetadata, *appError.AppError) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		return dto.AuthzMetadata{}, appError.NewAppError(http.StatusForbidden, "forbidden", []appError.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	pgID, err := services.ToPgUUID(userID.(string))
	if err != nil {
		return dto.AuthzMetadata{}, err
	}
	return dto.AuthzMetadata{
		UserID: pgID,
	}, nil
}
