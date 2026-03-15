package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appError "github.com/rijum8906/go-micro-service/packages/common/errors"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

func ExtractAuthzMeatadata(ctx *gin.Context) (dto.AuthzMetadata, *appError.AppError) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		return dto.AuthzMetadata{}, appError.NewAppError(http.StatusForbidden, "forbidden", []appError.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	pgID, err := StrIDToPgUUID(userID.(string))
	if err != nil {
		return dto.AuthzMetadata{}, err
	}
	return dto.AuthzMetadata{
		UserID: pgID,
	}, nil
}

func ExtractReqMetadata(deviceID string, ctx *gin.Context) dto.RequestMetadata {
	return dto.RequestMetadata{
		DeviceID:  deviceID,
		UserAgent: ctx.Request.UserAgent(),
		IPAddr:    ctx.ClientIP(),
	}
}
