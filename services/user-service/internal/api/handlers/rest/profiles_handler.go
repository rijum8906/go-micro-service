package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/response"
	"github.com/rijum8906/relay/services/user-service/internal/api/middleware"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/services/profile"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

func SetupProfilesHandlers(router *gin.RouterGroup, service profile.ProfileService, middlewareService middleware.Middleware) {
	// Public routes
	router.GET("/:profile_id", getProfile(service))

	// Private routes
	authorized := router.Group("/")
	authorized.Use(middlewareService.AuthMiddleware())
	{
		authorized.PUT("/:profile_id", updateProfile(service))
		authorized.DELETE("/:profile_id", deleteProfile(service))
		authorized.POST("/", createProfile(service))
	}
}

func getProfile(service profile.ProfileService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		profileID := ctx.Param("profile_id")
		result, err := service.GetProfile(ctx, profileID)
		if err != nil {
			utils.HandleServiceError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func createProfile(service profile.ProfileService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.CreateProfileRequest
		if err := ctx.ShouldBind(&input); err != nil {
			utils.HandleBindError(ctx, err)
			return
		}

		reqMetadata := utils.ExtractReqMetadata(input.Metadata.DeviceID, ctx)
		authMetadata, appErr := utils.ExtractAuthzMeatadata(ctx)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		result, appErr := service.CreateProfile(ctx, input, reqMetadata, authMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[*db.Profile]{
			Success: true,
			Message: "Profile created successfully",
			Data:    result,
		})
	}
}

func updateProfile(service profile.ProfileService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.UpdateProfileRequest
		profileID := ctx.Param("profile_id")
		input.ProfileID = profileID

		fileHeader, err := ctx.FormFile("avatar")
		if err == nil {
			input.Avatar = fileHeader
		} else {
			input.Avatar = nil
		}

		// 1. Use ShouldBind instead of ShouldBindJSON for multipart/form-data
		if err := ctx.ShouldBind(&input); err != nil {
			utils.HandleBindError(ctx, err)
			return
		}

		reqMetadata := utils.ExtractReqMetadata(input.Metadata.DeviceID, ctx)
		authMetadata, appErr := utils.ExtractAuthzMeatadata(ctx)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		result, appErr := service.UpdateProfile(ctx, input, reqMetadata, authMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[*db.Profile]{
			Success: true,
			Message: "Profile updated successfully",
			Data:    result,
		})
	}
}

func deleteProfile(service profile.ProfileService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		profileID := ctx.Param("profile_id")

		authMetadata, appErr := utils.ExtractAuthzMeatadata(ctx)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		appErr = service.DeleteProfile(ctx, profileID, authMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[*db.Profile]{
			Success: true,
			Message: "Profile deleted successfully",
		})
	}
}
