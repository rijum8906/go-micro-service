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
