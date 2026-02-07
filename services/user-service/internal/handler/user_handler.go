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

		if err := ctx.ShouldBindJSON(&input); err != nil {
			handleBindError(ctx, err)
			return
		}
		data := dto.UpdateProfileRequest{
			ProfileID: input.ProfileID,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			Metadata: dto.Metadata{
				DeviceID: input.Metadata.DeviceID,
			},
		}

		reqMetadata := dto.RequestMetadata{
			UserAgent: ctx.Request.UserAgent(),
			IPAddr:    ctx.ClientIP(),
			DeviceID:  input.Metadata.DeviceID,
		}

		userID, ok := ctx.Get("user_id")
		if !ok {
			ctx.JSON(http.StatusInternalServerError, &dto.BaseErrorResponse{
				Success: false,
				Message: "user id not found",
			})
			return
		}
		uidStr, ok := userID.(string)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, &dto.BaseErrorResponse{
				Success: false,
				Message: "user id not found",
			})
			return
		}
		pgUUID, err := services.ToPgUUID(uidStr)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, &dto.BaseErrorResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		authzMetadata := dto.AuthzMetadata{
			UserID: pgUUID,
		}

		err = service.UpdateProfile(ctx.Request.Context(), data, reqMetadata, authzMetadata)
		if err != nil {
			ctx.JSON(err.StatusCode, parseAppError(err))
			return
		}

		ctx.JSON(http.StatusOK, &dto.BaseSuccessResponse[*dto.AuthResponse]{
			Success: true,
			Message: "profile updated successfully",
		})
	}
}
