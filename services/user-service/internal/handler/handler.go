// Package handler contains HTTP handlers for the auth service.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services"
)

// RegisterHandlers sets up routes for the auth service.
func RegisterHandlers(router *gin.RouterGroup, service services.AuthService) {
	router.POST("/signup", signUpHandler(service))
	router.POST("/signin", signInHandler(service))
}

// --------------------
// Handlers
// --------------------

func signUpHandler(service services.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.SignUpDTO

		if err := ctx.ShouldBindJSON(&input); err != nil {
			handleBindError(ctx, err)
			return
		}

		result, err := service.SignUp(ctx.Request.Context(), input)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to create account",
			})
			return
		}

		ctx.JSON(http.StatusCreated, result)
	}
}

func signInHandler(service services.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.SignInDTO

		if err := ctx.ShouldBindJSON(&input); err != nil {
			handleBindError(ctx, err)
			return
		}

		result, err := service.Signin(ctx.Request.Context(), input)
		if err != nil {
			// Auth failure should be explicit and boring
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

// --------------------
// Error Handling
// --------------------

// handleBindError differentiates between validation errors and malformed JSON.
func handleBindError(ctx *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"validation_errors": formatValidationErrors(ve),
		})
		return
	}

	// Covers invalid JSON, wrong types, etc.
	ctx.JSON(http.StatusBadRequest, gin.H{
		"error": "malformed request body",
	})
}

func formatValidationErrors(ve validator.ValidationErrors) map[string]string {
	errorsMap := make(map[string]string)

	for _, fe := range ve {
		field := fe.Field()
		errorsMap[field] = validationErrorMessage(fe)
	}

	return errorsMap
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
