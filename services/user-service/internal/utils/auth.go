package utils

import (
	"github.com/gin-gonic/gin"
	appError "github.com/rijum8906/relay/packages/common/errors"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/api/middleware"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
)

func ExtractAuthzMeatadata(ctx *gin.Context) (request.AuthzMetadata, *appError.AppError) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return request.AuthzMetadata{}, appError.NewAppError(appError.ErrForbidden.Code, "forbidden", []appError.Error{
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

func ParseProfile(profile *db.Profile) *user_servicev1.Profile {
	return &user_servicev1.Profile{
		Id:          NewID(profile.ID.String()),
		FirstName:   NewName(profile.FirstName),
		LastName:    NewName(profile.LastName),
		DisplayName: NewName(profile.DisplayName.String),
		AvatarUrl:   NewURL(profile.AvatarUrl.String),
		CreatedAt:   NewTimestamp(profile.CreatedAt.Time),
		UpdatedAt:   NewTimestamp(profile.UpdatedAt.Time),
	}
}

func ParseAccount(account *db.Account) *user_servicev1.Account {
	return &user_servicev1.Account{
		Id:                 NewID(account.ID.String()),
		Email:              NewEmail(account.Email),
		IsEmailVerified:    account.IsEmailVerified,
		EmailVerifiedAt:    NewTimestamp(account.EmailVerifiedAt.Time),
		TwoFactorEnabled:   account.TwoFactorEnabled,
		TwoFactorEnabledAt: NewTimestamp(account.TwoFactorEnabledAt.Time),
		CreatedAt:          NewTimestamp(account.CreatedAt.Time),
		UpdatedAt:          NewTimestamp(account.UpdatedAt.Time),
	}
}
