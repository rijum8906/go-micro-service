// Package services contains services for the auth service.
package services

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	appError "github.com/rijum8906/go-micro-service/packages/common/errors"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/utils"
)

// ToPgUUID converts a string to pgtype.UUID or returns a professional 400 error.
// Useful for bridging between domain strings and database-specific types.
func ToPgUUID(idStr string) (pgtype.UUID, *appError.AppError) {
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		return pgtype.UUID{}, appError.NewAppError(http.StatusBadRequest, "bad request", []appError.Error{
			{Field: "id", Message: "The provided ID is not a valid UUID format."},
		})
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

// --- Authentication Logic ---

// Signin handles the full authentication flow: verification, session creation, and token issuance.
func (s *authService) Signin(ctx context.Context, data dto.SigninRequest, reqMetadata dto.RequestMetadata) (*dto.AuthResponse, *appError.AppError) {
	// 1. Verify Account Existence
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appError.ErrInvalidCredentials // Hide user existence for security
		}
		return nil, appError.ErrInternal.WithInternal(err)
	}

	// 2. Credential Verification
	if err := s.utilsConfig.HashService.VerifyPassword(account.PasswordHash, data.Password); err != nil {
		return nil, appError.ErrInvalidCredentials
	}

	// 3. Data Retrieval & Session Preparation
	profiles, err := s.q.GetProfilesByAccountID(ctx, account.ID)
	if err != nil {
		return nil, appError.ErrInternal.WithInternal(err)
	}

	refreshToken, err := s.utilsConfig.HashService.GenerateRefreshToken()
	if err != nil {
		return nil, appError.ErrInternal.WithInternal(err)
	}

	// 4. Persistence: Create the Refresh Session
	// Note: In a larger app, you might use a transaction here.
	createdSession, err := s.q.CreateSession(ctx, db.CreateSessionParams{
		AccountID:    account.ID,
		RefreshToken: refreshToken,
		UserAgent:    reqMetadata.UserAgent,
		IpAddr:       reqMetadata.IPAddr,
		DeviceID:     reqMetadata.DeviceID,
		ExpiresAt:    pgtype.Timestamptz{Time: time.Now().Add(s.env.JwtExpiration), Valid: true},
	})
	if err != nil {
		return nil, appError.ErrInternal.WithInternal(err)
	}

	// 5. Token Generation
	accessToken, err2 := s.utilsConfig.JwtService.IssueToken(ctx, utils.GenerateRedisLoginKey(account.ID.String(), createdSession.DeviceID))
	if err2 != nil {
		return nil, appError.ErrInternal.WithInternal(err2)
	}

	return &dto.AuthResponse{
		Account:  &account,
		Profiles: &profiles,
		Token:    &dto.Token{AccessToken: accessToken, RefreshToken: refreshToken},
	}, nil
}

// SignUp orchestrates account creation and immediate profile initialization.
func (s *authService) SignUp(ctx context.Context, data dto.SignupRequest, reqMetadata dto.RequestMetadata) (*dto.AuthResponse, *appError.AppError) {
	// 1. Idempotency/Duplicate Check
	_, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err == nil {
		return nil, appError.ErrUserExists.WithField("email", "An account with this email already exists.")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, appError.ErrInternal.WithInternal(err)
	}

	// 2. Security: Securely Hash Password
	passHash, err := s.utilsConfig.HashService.HashPassword(data.Password)
	if err != nil {
		return nil, appError.ErrInternal.WithInternal(err)
	}

	// 3. Database Write Operations
	// TODO: Wrap in a DB transaction to ensure Account + Profile are atomic.
	account, err := s.q.CreateAccount(ctx, db.CreateAccountParams{
		Email:        data.Email,
		PasswordHash: passHash,
	})
	if err != nil {
		return nil, appError.ErrInternal.WithInternal(err)
	}

	profile, err := s.q.CreateProfile(ctx, db.CreateProfileParams{
		AccountID: account.ID,
		FirstName: data.FirstName,
		LastName:  data.LastName,
	})
	if err != nil {
		return nil, appError.ErrInternal.WithInternal(err)
	}

	// 4. Post-Registration: Auto-Login session
	refreshToken, err := s.utilsConfig.HashService.GenerateRefreshToken()
	if err != nil {
		return nil, appError.ErrInternal.WithInternal(err)
	}
	createdSession, err := s.q.CreateSession(ctx, db.CreateSessionParams{
		AccountID:    account.ID,
		RefreshToken: refreshToken,
		UserAgent:    reqMetadata.UserAgent,
		IpAddr:       reqMetadata.IPAddr,
		DeviceID:     reqMetadata.DeviceID,
		ExpiresAt:    pgtype.Timestamptz{Time: time.Now().Add(s.env.JwtExpiration), Valid: true},
	})
	if err != nil {
		return nil, appError.ErrInternal.WithInternal(err)
	}

	accessToken, appErr := s.utilsConfig.JwtService.IssueToken(ctx, utils.GenerateRedisLoginKey(account.ID.String(), createdSession.DeviceID))
	if appErr != nil {
		return nil, appError.ErrInternal.WithInternal(appErr)
	}

	return &dto.AuthResponse{
		Account:  &account,
		Profiles: &[]db.Profile{profile},
		Token:    &dto.Token{AccessToken: accessToken, RefreshToken: refreshToken},
	}, nil
}

// --- Account Maintenance Logic ---

// RequestPasswordReset sends a secure link. Note: We return "NotFound" specifically
// here because reset requests are usually initiated by users who know their email.
func (s *authService) RequestPasswordReset(ctx context.Context, data dto.RequestPasswordResetRequest, reqMetadata dto.RequestMetadata) *appError.AppError {
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appError.ErrNotFound
		}
		return appError.ErrInternal.WithInternal(err)
	}

	// Create a short-lived scoped token for the reset action
	_, appErr := s.utilsConfig.JwtService.IssueToken(ctx, utils.FormatUUID(account.ID))
	if appErr != nil {
		return appErr
	}

	return nil
}

// ResetPassword consumes a token to update credentials.
func (s *authService) ResetPassword(ctx context.Context, data dto.ResetPasswordRequest, reqMetadata dto.RequestMetadata) *appError.AppError {
	// 1. Verify Action Token Integrity
	claims, appErr := s.utilsConfig.JwtService.ValidateToken(ctx, data.Token)
	if appErr != nil {
		return appErr
	}

	// 2. Hash New Credentials
	hashPass, err := s.utilsConfig.HashService.HashPassword(data.NewPassword)
	if err != nil {
		return appError.ErrInternal.WithInternal(err)
	}

	// 3. Map Domain ID to DB ID
	pgUUID, appErr := ToPgUUID(claims.UserID)
	if appErr != nil {
		return appErr
	}

	// 4. Commit Changes
	if _, err = s.q.UpdateAccount(ctx, db.UpdateAccountParams{ID: pgUUID, PasswordHash: pgtype.Text{
		String: hashPass,
		Valid:  true,
	}}); err != nil {
		return appError.ErrInternal.WithInternal(err)
	}
	return nil
}

// VerifyEmail marks an account as trusted after validating the provided token.
func (s *authService) VerifyEmail(ctx context.Context, data dto.VerifyEmailRequest, reqMetadata dto.RequestMetadata) *appError.AppError {
	claims, appErr := s.utilsConfig.JwtService.ValidateToken(ctx, data.Token)
	if appErr != nil {
		return appErr
	}

	pgUUID, appErr := ToPgUUID(claims.UserID)
	if appErr != nil {
		return appErr
	}

	// Update verification status and timestamp
	_, err := s.q.UpdateAccountSecurityByAccountID(ctx, db.UpdateAccountSecurityByAccountIDParams{
		IsEmailVerified: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
		AccountID:       pgUUID,
		EmailVerifiedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
	})
	if err != nil {
		return appError.ErrInternal.WithInternal(err)
	}

	return nil
}

func (s *authService) RequestEmailVerification(ctx context.Context, data dto.RequestEmailVerificationRequest, reqMetadata dto.RequestMetadata) *appError.AppError {
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appError.ErrNotFound
		}
		return appError.ErrInternal.WithInternal(err)
	}

	sec, err := s.q.GetAccountSecurityByAccountID(ctx, account.ID)
	if err != nil {
		return appError.ErrInternal.WithInternal(err)
	}

	if sec.IsEmailVerified {
		return appError.NewAppError(http.StatusBadRequest, "email already verified", []appError.Error{})
	}

	if _, appErr := s.utilsConfig.JwtService.IssueToken(ctx, utils.FormatUUID(account.ID)); appErr != nil {
		return appErr
	}

	return nil
}
