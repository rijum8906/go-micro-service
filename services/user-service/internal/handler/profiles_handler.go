package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/middleware"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/utils"
)

func SetupProfilesHandlers(router *gin.RouterGroup, service services.ProfileService, middlewareService middleware.Middleware) {
	router.Use(func(ctx *gin.Context) {
		if ctx.Request.Method != "GET" {
			middlewareService.AuthMiddleware()(ctx)
			return
		}
		ctx.Next()
	})
	router.GET("/:profile_id", getProfile(service))
	router.PUT("/:profile_id", updateProfile(service))
	router.DELETE("/:profile_id", deleteProfile(service))
	router.POST("/", createProfile(service))
}

func getProfile(service services.ProfileService) gin.HandlerFunc {
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

func createProfile(service services.ProfileService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.CreateProfileRequest
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
		ctx.JSON(http.StatusOK, dto.BaseSuccessResponse[*db.Profile]{
			Success: true,
			Message: "Profile created successfully",
			Data:    result,
		})
	}
}

func updateProfile(service services.ProfileService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.UpdateProfileRequest
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
		ctx.JSON(http.StatusOK, dto.BaseSuccessResponse[*db.Profile]{
			Success: true,
			Message: "Profile updated successfully",
			Data:    result,
		})
	}
}

func deleteProfile(service services.ProfileService) gin.HandlerFunc {
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
		ctx.JSON(http.StatusOK, dto.BaseSuccessResponse[*db.Profile]{
			Success: true,
			Message: "Profile deleted successfully",
		})
	}
}
