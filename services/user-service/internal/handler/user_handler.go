package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/middleware"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services"
)

// RegisterUserHandlers sets up routes for the auth service.
func RegisterUserHandlers(router *gin.RouterGroup, service services.UserService, middlewraeService middleware.Middleware) {
	router.Use(middlewraeService.AuthMiddleware())
	router.PUT("/update-profile", updateProfile(service))
}

func updateProfile(service services.UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.UpdateProfileRequest

		// 1. Use ShouldBind instead of ShouldBindJSON for multipart/form-data
		if err := ctx.ShouldBind(&input); err != nil {
			handleBindError(ctx, err)
			return
		}

		// 2. Validate user_id from middleware
		userID, ok := ctx.Get("user_id")
		if !ok {
			ctx.JSON(http.StatusUnauthorized, &dto.BaseErrorResponse{
				Success: false,
				Message: "user id not found",
			})
			return
		}

		uidStr, _ := userID.(string)
		pgUUID, appErr := services.ToPgUUID(uidStr)
		if appErr != nil {
			ctx.JSON(appErr.StatusCode, parseAppError(appErr))
			return
		}

		// 3. Prepare Metadata
		reqMetadata := dto.RequestMetadata{
			UserAgent: ctx.Request.UserAgent(),
			IPAddr:    ctx.ClientIP(),
			DeviceID:  input.Metadata.DeviceID,
		}

		authzMetadata := dto.AuthzMetadata{
			UserID: pgUUID,
		}

		// 4. Pass the entire 'input' struct (which now contains input.Avatar)
		appErr = service.UpdateProfile(ctx.Request.Context(), input, reqMetadata, authzMetadata)
		if appErr != nil {
			ctx.JSON(appErr.StatusCode, parseAppError(appErr))
			return
		}

		ctx.JSON(http.StatusOK, &dto.BaseSuccessResponse[any]{
			Success: true,
			Message: "profile updated successfully",
		})
	}
}
