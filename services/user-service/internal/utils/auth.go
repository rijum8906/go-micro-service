package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appError "github.com/rijum8906/go-micro-service/packages/common/errors"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/middleware"
)

func ExtractAuthzMeatadata(ctx *gin.Context) (request.AuthzMetadata, *appError.AppError) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return request.AuthzMetadata{}, appError.NewAppError(http.StatusForbidden, "forbidden", []appError.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	pgID, err := StrIDToPgUUID(userID)
	if err != nil {
		return request.AuthzMetadata{}, err
	}
	return request.AuthzMetadata{
		UserID: pgID,
	}, nil
}

func ExtractReqMetadata(deviceID string, ctx *gin.Context) request.RequestMetadata {
	return request.RequestMetadata{
		DeviceID:  deviceID,
		UserAgent: ctx.Request.UserAgent(),
		IPAddr:    ctx.ClientIP(),
	}
}
