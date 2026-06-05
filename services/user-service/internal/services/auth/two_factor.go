package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func (s *AuthService) EnableTwoFactor(ctx context.Context, req *corev1.EmptyRequest) (*authv1.EnableTwoFactorResponse, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	var totpURI string

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		user, err := s.DBQ.GetUser(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrNotFound.WithMessage("user not found")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to get user").
				WithDetail("db_error", err.Error())
		}

		secret, appErr := s.HashService.GenerateBase32(32)
		if appErr != nil {
			return appErr
		}

		totpURI = utils.GenerateTOTPAuthURI(utils.TOTPAuthURIConfig{
			Protocol:     utils.TOTPProtocolGoogle,
			Type:         utils.TOTPTypeTOTP,
			Issuer:       s.Config.AppName,
			Email:        user.Email,
			Secret:       secret,
			Algorithm:    "SHA1",
			CodeLength:   6,
			CodeValidity: time.Minute * 30,
		})

		if err = s.DBQ.EnableUserTwoFactor(ctx, db.EnableUserTwoFactorParams{
			ID: userID,
			TwoFactorSecret: pgtype.Text{
				String: secret,
				Valid:  true,
			},
		}); err != nil {
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to enable user two factor").
				WithDetail("db_error", err.Error())
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	return &authv1.EnableTwoFactorResponse{
		TotpUri: totpURI,
	}, nil
}

func (s *AuthService) DisableTwoFactor(ctx context.Context, req *corev1.EmptyRequest) (*corev1.SuccessResponse, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	if err = s.DBQ.DisableUserTwoFactor(ctx, userID); err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to disable user two factor").
			WithDetail("db_error", err.Error())
	}

	return &corev1.SuccessResponse{Success: true}, nil
}
