package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/response"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/middleware"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/services/account"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/utils"
)

func SetupAccountsHandlers(router *gin.RouterGroup, service account.AccountService, middlewareService middleware.Middleware) {
	// Public routes
	router.GET("/:id/exists", checkAccountExist(service))

	// Private routes
	authorized := router.Group("/")
	authorized.Use(middlewareService.AuthMiddleware())
	{
		authorized.POST("/signout", signOutHandler(service))
		authorized.PUT("/change-email", changeEmail(service))
		authorized.PUT("/change-password", changePassword(service))
		authorized.DELETE("/:id", deleteAccount(service))
		authorized.GET("/my-account", myAccount(service))
		authorized.POST("/generate-scoped-token", generateScopedToken(service))
	}
}

func myAccount(service account.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.Request
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

func deleteAccount(service account.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		pgID, appErr := utils.StrIDToPgUUID(id)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		var input request.Request
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
			ctx.JSON(http.StatusForbidden, response.BaseErrorResponse{
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
		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[bool]{
			Success: true,
			Message: "Account deleted successfully",
		})
	}
}

func checkAccountExist(service account.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		pgID, appErr := utils.StrIDToPgUUID(id)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}
		result, appErr := service.CheckAccountExist(ctx, pgID)
		if appErr != nil {
			utils.HandleServiceError(ctx, appErr)
			return
		}

		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[*response.CheckAccountExistResult]{
			Success: true,
			Message: "Account existence checked successfully",
			Data:    result,
		})
	}
}

func changeEmail(service account.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.ChangeEmailRequest
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

		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[*response.ChangeEmailResult]{
			Success: true,
			Message: "Email changed successfully",
			Data:    result,
		})
	}
}

func changePassword(service account.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.ChangePasswordRequest
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
		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[bool]{
			Success: true,
			Message: "Password changed successfully",
		})
	}
}

func generateScopedToken(service account.AccountService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input request.GenerateScopedTokenRequest
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
		ctx.JSON(http.StatusOK, response.BaseSuccessResponse[*response.GenerateScopedTokenResult]{
			Success: true,
			Message: "Scoped token generated successfully",
			Data:    result,
		})
	}
}

func signOutHandler(service account.AccountService) gin.HandlerFunc {
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
