// Package handler contains HTTP handlers for the auth service.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/response"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services/auth"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/utils"
)

// RegisterHandlers sets up routes for the auth service.
func RegisterHandlers(router *gin.RouterGroup, service auth.AuthService) {
	router.POST("/signup", signUpHandler(service))
	router.POST("/signin", signInHandler(service))
}

// --------------------
// Handlers
// --------------------

func signUpHandler(service auth.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.SignupRequest

		if err := ctx.ShouldBindJSON(&input); err != nil {
			utils.HandleBindError(ctx, err)
			return
		}

		requestMetadata := request.RequestMetadata{
			UserAgent: ctx.Request.UserAgent(),
			IPAddr:    ctx.ClientIP(),
			DeviceID:  input.Metadata.DeviceID,
		}

		result, err := service.SignUp(ctx.Request.Context(), input, requestMetadata)
		if err != nil {
			utils.HandleServiceError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, &response.BaseSuccessResponse[*response.AuthResponse]{
			Success: true,
			Message: "account created successfully",
			Data:    result,
		})
	}
}

func signInHandler(service auth.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.SigninRequest

		if err := ctx.ShouldBindJSON(&input); err != nil {
			utils.HandleBindError(ctx, err)
			return
		}

		requestMetadata := request.RequestMetadata{
			UserAgent: ctx.Request.UserAgent(),
			IPAddr:    ctx.ClientIP(),
			DeviceID:  input.Metadata.DeviceID,
		}

		result, err := service.Signin(ctx.Request.Context(), input, requestMetadata)
		if err != nil {
			utils.HandleServiceError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, &response.BaseSuccessResponse[*response.AuthResponse]{
			Success: true,
			Message: "account signed in successfully",
			Data:    result,
		})
	}
}
