package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/middleware"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/utils"
)

func SetupAccountsHandlers(router *gin.RouterGroup, service services.AccountService, middlewareService middleware.Middleware) {
	router.Use(func(ctx *gin.Context) {
		if ctx.Request.Method != "GET" {
			middlewareService.AuthMiddleware()(ctx)
			return
		}
		ctx.Next()
	})

	router.POST("/signout", signOutHandler(service))
	router.PUT("/change-email", changeEmail(service))
	router.PUT("/change-password", changePassword(service))
	router.DELETE("/:id", deleteAccount(service))
	router.GET("/:id/exists", checkAccountExist(service))
	router.GET("/my-account", myAccount(service))
	router.POST("/generate-scoped-token", generateScopedToken(service))
}

func myAccount(service services.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.Request
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

		account, appErr := service.MyAccount(ctx, reqMetadata, authMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		ctx.JSON(http.StatusOK, account)
	}
}

func deleteAccount(service services.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		pgID, appErr := services.ToPgUUID(id)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		var input dto.Request
		if err := ctx.ShouldBindJSON(&input); err != nil {
			utils.HandleBindError(ctx, appErr)
			return
		}

		reqMetadata := utils.ExtractReqMetadata(input.Metadata.DeviceID, ctx)
		authMetadata, appErr := utils.ExtractAuthzMeatadata(ctx)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		if authMetadata.UserID != pgID {
			ctx.JSON(http.StatusForbidden, dto.BaseErrorResponse{
				Success: false,
				Message: "You do not have permission to perform this action.",
			})
			return
		}

		appErr = service.DeleteAccount(ctx, reqMetadata, authMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		ctx.JSON(http.StatusOK, dto.BaseSuccessResponse[bool]{
			Success: true,
			Message: "Account deleted successfully",
		})
	}
}

func checkAccountExist(service services.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		pgID, appErr := services.ToPgUUID(id)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		result, appErr := service.CheckAccountExist(ctx, pgID)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		ctx.JSON(http.StatusOK, dto.BaseSuccessResponse[*dto.CheckAccountExistResult]{
			Success: true,
			Message: "Account existence checked successfully",
			Data:    result,
		})
	}
}

func changeEmail(service services.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.ChangeEmailRequest
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

		result, appErr := service.ChangeEmail(ctx, input, reqMetadata, authMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		ctx.JSON(http.StatusOK, dto.BaseSuccessResponse[*dto.ChangeEmailResult]{
			Success: true,
			Message: "Email changed successfully",
			Data:    result,
		})
	}
}

func changePassword(service services.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.ChangePasswordRequest
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

		appErr = service.ChangePassword(ctx, input, reqMetadata, authMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		ctx.JSON(http.StatusOK, dto.BaseSuccessResponse[bool]{
			Success: true,
			Message: "Password changed successfully",
		})
	}
}

func generateScopedToken(service services.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.GenerateScopedTokenRequest
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

		result, appErr := service.GenerateScopedToken(ctx, input, reqMetadata, authMetadata)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		ctx.JSON(http.StatusOK, dto.BaseSuccessResponse[*dto.GenerateScopedTokenResult]{
			Success: true,
			Message: "Scoped token generated successfully",
			Data:    result,
		})
	}
}

func signOutHandler(service services.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input dto.SignoutRequest
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
		ctx.JSON(http.StatusOK, dto.BaseSuccessResponse[*dto.GenerateScopedTokenResult]{
			Success: true,
			Message: "Signed out successfully",
		})
	}
}
