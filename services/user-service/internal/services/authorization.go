package services

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	appError "github.com/rijum8906/go-micro-service/packages/common/errors"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

func (s *authService) GenerateScopedActionToken(ctx context.Context, data dto.GenerateScopedTokenRequest, authzMetadata dto.AuthzMetadata) (string, *appError.AppError) {
	token, appErr := s.utilsConfig.SecureJWTService.IssueToken(ctx, jwt.ScopedActionClaims{
		Scope:   data.Scope,
		Subject: authzMetadata.UserID.String(),
	})
	if appErr != nil {
		return "", appError.ErrInternal.WithInternal(appErr)
	}
	return token, nil
}

func (s *authService) ChangePassword(ctx context.Context, data dto.ChangePasswordRequest, authzMetadata dto.AuthzMetadata) *appError.AppError {
	claims, appErr := s.utilsConfig.SecureJWTService.ValidateToken(ctx, data.Token)
	if appErr != nil {
		return appError.ErrInvalidToken
	}

	// Professional Check: Ensure the token belongs to the acting user
	if claims.Subject != authzMetadata.UserID.String() {
		return appError.NewAppError(http.StatusForbidden, "forbidden", []appError.Error{
			{Field: "auth", Message: "You do not have permission to perform this action."},
		})
	}

	newPassHash, err := s.utilsConfig.HashService.HashPassword(data.NewPassword)
	if err != nil {
		return appError.ErrInternal.WithInternal(err)
	}

	_, err = s.q.UpdateAccount(ctx, db.UpdateAccountParams{
		ID: authzMetadata.UserID,
		PasswordHash: pgtype.Text{
			String: newPassHash,
			Valid:  true,
		},
	})
	if err != nil {
		return appError.ErrInternal.WithInternal(err)
	}

	return nil
}
