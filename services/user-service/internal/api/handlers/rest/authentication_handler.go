// Package handler contains HTTP handlers for the auth service.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/response"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/middleware"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services/auth"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/utils"
)

// RegisterHandlers sets up routes for the auth service.
func RegisterHandlers(router *gin.RouterGroup, service auth.AuthService, middlewareService middleware.Middleware) {
	router.POST("/signup", signUpHandler(service))
	router.POST("/signin", signInHandler(service))
	router.POST("/signout", middlewareService.AuthMiddleware(), signOutHandler(service))
	router.POST("/request-eamil-verificaton", requestEmailVerificationHandler(service))
	router.POST("/verify-email", verifyEmailHandler(service))
	router.POST("/request-password-reset", requestPasswordResetHandler(service))
	router.POST("/reset-password", resetPasswordHandler(service))
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

func signOutHandler(service auth.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.SignoutRequest
		if err := ctx.ShouldBindJSON(&input); err != nil {
			utils.HandleBindError(ctx, err)
			return
		}

		reqMetadata := utils.ExtractReqMetadata(input.Metadata.DeviceID, ctx)
		authMetadata, appErr := utils.ExtractAuthzMeatadata(ctx)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		appErr = service.Signout(ctx, reqMetadata, authMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[*response.GenerateScopedTokenResult]{
			Success: true,
			Message: "Signed out successfully",
		})
	}
}

func requestEmailVerificationHandler(service auth.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.RequestEmailVerificationRequest
		if err := ctx.ShouldBindJSON(&input); err != nil {
			utils.HandleBindError(ctx, err)
			return
		}

		reqMetadata := utils.ExtractReqMetadata(input.Metadata.DeviceID, ctx)

		appErr := service.RequestEmailVerification(ctx, input, reqMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[any]{
			Success: true,
			Message: "Email verification request sent successfully",
		})
	}
}

func verifyEmailHandler(service auth.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.VerifyEmailRequest
		if err := ctx.ShouldBindJSON(&input); err != nil {
			utils.HandleBindError(ctx, err)
			return
		}

		reqMetadata := utils.ExtractReqMetadata(input.Metadata.DeviceID, ctx)

		appErr := service.VerifyEmail(ctx, input, reqMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[any]{
			Success: true,
			Message: "Email verified successfully",
		})
	}
}

func requestPasswordResetHandler(service auth.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.RequestPasswordResetRequest
		if err := ctx.ShouldBindJSON(&input); err != nil {
			utils.HandleBindError(ctx, err)
			return
		}

		reqMetadata := utils.ExtractReqMetadata(input.Metadata.DeviceID, ctx)

		appErr := service.RequestPasswordReset(ctx, input, reqMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[any]{
			Success: true,
			Message: "Password reset request sent successfully",
		})
	}
}

func resetPasswordHandler(service auth.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.ResetPasswordRequest
		if err := ctx.ShouldBindJSON(&input); err != nil {
			utils.HandleBindError(ctx, err)
			return
		}

		reqMetadata := utils.ExtractReqMetadata(input.Metadata.DeviceID, ctx)

		appErr := service.ResetPassword(ctx, input, reqMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[any]{
			Success: true,
			Message: "Password reset successfully",
		})
	}
}
