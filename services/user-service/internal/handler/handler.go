// Package handler contains HTTP handlers for the user service.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services"
)

// RegisterHandlers sets up authentication routes for the user service.
func RegisterHandlers(router *gin.RouterGroup, service services.AuthService) {
	router.POST("/signup", signUpHandler(service))
	router.POST("/signin", signInHandler(service))
}

func signUpHandler(service services.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.SignupRequest

		if err := ctx.ShouldBindJSON(&input); err != nil {
			handleBindError(ctx, err)
			return
		}
		result, err := service.SignUp(ctx.Request.Context(), input, buildRequestMetadata(ctx, input.Metadata.DeviceID))
		if err != nil {
			ctx.JSON(err.StatusCode, parseAppError(err))
			return
		}

		ctx.JSON(http.StatusOK, &dto.BaseSuccessResponse[*dto.AuthResponse]{
			Success: true,
			Message: "account created successfully",
			Data:    result,
		})
	}
}

func signInHandler(service services.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.SigninRequest

		if err := ctx.ShouldBindJSON(&input); err != nil {
			handleBindError(ctx, err)
			return
		}
		result, err := service.Signin(ctx.Request.Context(), input, buildRequestMetadata(ctx, input.Metadata.DeviceID))
		if err != nil {
			ctx.JSON(err.StatusCode, parseAppError(err))
			return
		}

		ctx.JSON(http.StatusOK, &dto.BaseSuccessResponse[*dto.AuthResponse]{
			Success: true,
			Message: "account signed in successfully",
			Data:    result,
		})
	}
}

func buildRequestMetadata(ctx *gin.Context, deviceID string) dto.RequestMetadata {
	return dto.RequestMetadata{
		UserAgent: ctx.Request.UserAgent(),
		IPAddr:    ctx.ClientIP(),
		DeviceID:  deviceID,
	}
}
