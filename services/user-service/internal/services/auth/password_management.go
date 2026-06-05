package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
)

func (s *AuthService) ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("change password request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	claims, appErr := s.TokenManager.ValidateScopedToken(ctx, req.TokenScope)
	if appErr != nil {
		return nil, appErr
	}
	if claims.Subject != userInfo.UserID {
		return nil, apperror.ErrValidation.WithMessage("invalid user id in token")
	}

	if claims.Scope != constants.TokenScopeChangePassword {
		return nil, apperror.ErrValidation.WithMessage("invalid scoped token scope for change password")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	newPasswordHash, appErr := s.HashService.Hash(req.NewPassword)
	if appErr != nil {
		return nil, appErr
	}

	if err = s.DBQ.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID: userID,
		PasswordHash: pgtype.Text{
			String: newPasswordHash,
			Valid:  true,
		},
	}); err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to update user").WithDetail("error", err.Error())
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}
