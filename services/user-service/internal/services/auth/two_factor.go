package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

// ============================================================
// TOTP
// ============================================================

// InitTwoFactorTOTP initializes two-factor authentication for a user using TOTP
//
// Business Logic:
//   - Generates a TOTP URI and secret for the user
//
// Idempotent: Yes
//
// Returns:
//   - The TOTP URI and secret for the user
func (s *AuthService) InitTwoFactorTOTP(ctx context.Context, req *corev1.EmptyRequest) (*authv1.InitTwoFactorTOTPResponse, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	var (
		totpURI string
		secret  string
	)

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

		twoFactorSecret, appErr := s.HashService.GenerateBase32(32)
		if appErr != nil {
			return appErr
		}
		secret = twoFactorSecret

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

		if err = s.DBQ.SetTwoFactorSecretByUserID(ctx, db.SetTwoFactorSecretByUserIDParams{
			UserID: userID,
			Secret: secret,
		}); err != nil {
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to enable user two factor").
				WithDetail("db_error", err.Error())
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	return &authv1.InitTwoFactorTOTPResponse{
		TwoFactorSecret: secret,
		QrCodeUri:       totpURI,
	}, nil
}

// EnableTwoFactorTOTP enables two-factor authentication for a user using TOTP
//
// Business Logic:
//   - Verifies the provided TOTP code against the stored secret
//   - Updates the user's two-factor authentication status in the database
//
// Idempotent: Yes
//
// Returns:
//   - A success response indicating the two-factor authentication was enabled
func (s *AuthService) EnableTwoFactorTOTP(ctx context.Context, req *authv1.TwoFactorTOTPRequest) (*corev1.SuccessResponse, error) {
	// Validate request
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request is nil")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, constants.ErrUserNotFoundInCtx
	}

	// Validate two factor code
	if appErr := validate2FACode(req.GetTotp()); appErr != nil {
		return nil, appErr
	}

	// Get User from database
	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, constants.ErrInvalidUserIDInUserInfo
	}

	// Check if the user has two-factor authentication enabled for TOTP
	exists, err := s.DBQ.CheckTwoFactorAuthEnabledByUserIDAndMethod(ctx, db.CheckTwoFactorAuthEnabledByUserIDAndMethodParams{
		UserID: userID,
		Method: constants.TwoFactorMethodTotp,
	})
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to retrieve user from database").
			WithDetail("db_error", err.Error())
	}
	if exists {
		return nil, apperror.ErrConflict.WithMessage("two factor auth already enabled for TOTP")
	}

	// Generate and Match two factor code against generated code
	generatedTwoFactorCode, appErr := generate2FATokenCode(req.TwoFactorSecret)
	if appErr != nil {
		return nil, appErr
	}
	if req.GetTotp() != generatedTwoFactorCode {
		return nil, apperror.ErrValidation.WithMessage("invalid two factor code")
	}

	// Enable two factor for user
	if err = s.DBQ.SetTwoFactorSecretByUserID(ctx, db.SetTwoFactorSecretByUserIDParams{
		UserID: userID,
		Secret: req.TwoFactorSecret,
	}); err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to enable user two factor").
			WithDetail("db_error", err.Error())
	}

	// TODO: Save to auth log
	return &corev1.SuccessResponse{Success: true}, nil
}

// DisableTwoFactor disables two-factor authentication for a user
//
// Business Logic:
//   - Removes the two-factor secret from the user's record in the database
//
// Idempotent: Yes
//
// Returns:
//   - A success response indicating the two-factor authentication was disabled
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

	exists, err := s.DBQ.CheckTwoFactorAuthEnabledByUserIDAndMethod(ctx, db.CheckTwoFactorAuthEnabledByUserIDAndMethodParams{
		UserID: userID,
		Method: constants.TwoFactorMethodTotp,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to check two factor auth enabled").WithDetail("db_error", err.Error())
	}
	if !exists {
		return nil, apperror.ErrValidation.WithMessage("two factor auth not enabled").WithDetail("method", constants.TwoFactorMethodTotp)
	}

	if err = s.DBQ.DisableTwoFactorAuthByUserIDAndMethod(ctx, db.DisableTwoFactorAuthByUserIDAndMethodParams{
		UserID: userID,
		Method: constants.TwoFactorMethodTotp,
	}); err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to disable user two factor").
			WithDetail("db_error", err.Error())
	}

	return &corev1.SuccessResponse{Success: true}, nil
}
