package services

import (
	"context"

	"github.com/google/uuid"
	appError "github.com/rijum8906/go-micro-service/packages/common/errors"
	"github.com/rijum8906/go-micro-service/packages/common/jwt"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

func (s *authService) GenerateScopedActionToken(ctx context.Context, data dto.GenerateScopedTokenRequest, authzMetadata dto.AuthzMetadata) (string, error) {
	token, err := s.utilsConfig.SecureJWTService.IssueToken(ctx, jwt.ScopedActionClaims{
		UserID: authzMetadata.UserID.String(),
		Scope:  data.Scope,
		JTI:    uuid.New().String(),
	})
	if err != nil {
		return "", appError.ErrInternal
	}
	return token, nil
}

func (s *authService) ChangePassword(ctx context.Context, data dto.ChangePasswordRequest, authzMetadata dto.AuthzMetadata) error {
	claims, err := s.utilsConfig.SecureJWTService.ValidateToken(ctx, data.Token)
	if err != nil {
		return appError.ErrInvalidToken
	}
	if claims.UserID != authzMetadata.UserID.String() {
		return appError.ErrInvalidTokenClaims
	}

	newPassHash, err := s.utilsConfig.HashService.HashPassword(data.NewPassword)
	if err != nil {
		return appError.ErrInternal
	}

	_, err = s.q.UpdateAccount(ctx, db.UpdateAccountParams{
		ID:           authzMetadata.UserID,
		PasswordHash: newPassHash,
	})
	if err != nil {
		return appError.ErrInternal
	}

	return nil
}
