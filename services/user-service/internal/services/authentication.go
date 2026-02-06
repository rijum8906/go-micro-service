// Package services
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
)

// Professional Global Error Definitions
var (
	ErrInternal = appError.NewAppError(http.StatusInternalServerError, "internal server error", &[]appError.Error{
		{Field: "server", Message: "An unexpected error occurred. Please try again later."},
	})
	ErrInvalidToken = appError.NewAppError(http.StatusUnauthorized, "unauthorized", &[]appError.Error{
		{Field: "token", Message: "Your session has expired or the token is invalid."},
	})
	ErrInvalidCredentials = appError.NewAppError(http.StatusUnauthorized, "invalid credentials", &[]appError.Error{
		{Field: "auth", Message: "The email or password provided is incorrect."},
	})
	ErrEmailAlreadyExists = appError.NewAppError(http.StatusConflict, "conflict", &[]appError.Error{
		{Field: "email", Message: "An account with this email already exists."},
	})
	ErrAccountNotFound = appError.NewAppError(http.StatusNotFound, "not found", &[]appError.Error{
		{Field: "email", Message: "No account found with this email address."},
	})
	ErrEmailAlreadyVerified = appError.NewAppError(http.StatusBadRequest, "bad request", &[]appError.Error{
		{Field: "email", Message: "This email is already verified."},
	})
)

// ToPgUUID converts a string to pgtype.UUID or returns a professional 400 error
func ToPgUUID(idStr string) (pgtype.UUID, *appError.AppError) {
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		return pgtype.UUID{}, appError.NewAppError(http.StatusBadRequest, "bad request", &[]appError.Error{
			{Field: "id", Message: "The provided ID is not a valid UUID format."},
		})
	}
	return pgtype.UUID{
		Bytes: parsed,
		Valid: true,
	}, nil
}

// Signin handles user authentication with 401 status for security
func (s *authService) Signin(ctx context.Context, data dto.SigninRequest, reqMetadata dto.RequestMetadata) (*dto.AuthResponse, *appError.AppError) {
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials // Secure: Don't leak user existence
		}
		return nil, ErrInternal
	}

	if err := s.utilsConfig.HashService.VerifyPassword(account.PasswordHash, data.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	profiles, err := s.q.GetProfilesByAccountID(ctx, account.ID)
	if err != nil {
		return nil, ErrInternal
	}

	refreshToken, err := s.utilsConfig.HashService.GenerateRefreshToken()
	accessToken, err2 := s.utilsConfig.JwtService.IssueToken(ctx, FormatUUID(account.ID))
	if err != nil || err2 != nil {
		return nil, ErrInternal
	}

	return &dto.AuthResponse{
		Account:  &account,
		Profiles: &profiles,
		Token: &dto.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

// SignUp handles user registration with 409 Conflict check
func (s *authService) SignUp(ctx context.Context, data dto.SignupRequest, reqMetadata dto.RequestMetadata) (*dto.AuthResponse, *appError.AppError) {
	_, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err == nil {
		return nil, ErrEmailAlreadyExists // Professional 409
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInternal
	}

	passHash, err := s.utilsConfig.HashService.HashPassword(data.Password)
	if err != nil {
		return nil, ErrInternal
	}

	account, err := s.q.CreateAccount(ctx, db.CreateAccountParams{
		Email:        data.Email,
		PasswordHash: passHash,
	})
	if err != nil {
		return nil, ErrInternal
	}

	profile, err := s.q.CreateProfile(ctx, db.CreateProfileParams{
		FirstName: data.FirstName,
		LastName:  data.LastName,
	})
	if err != nil {
		return nil, ErrInternal
	}

	refreshToken, _ := s.utilsConfig.HashService.GenerateRefreshToken()
	accessToken, _ := s.utilsConfig.JwtService.IssueToken(ctx, FormatUUID(account.ID))

	return &dto.AuthResponse{
		Account:  &account,
		Profiles: &[]db.Profile{profile},
		Token: &dto.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, data dto.RequestPasswordResetRequest, reqMetadata dto.RequestMetadata) *appError.AppError {
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAccountNotFound
		}
		return ErrInternal
	}

	_, err = s.utilsConfig.JwtService.IssueToken(ctx, FormatUUID(account.ID))
	if err != nil {
		return ErrInternal
	}

	// TODO: Trigger Email Service logic here
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, data dto.ResetPasswordRequest, reqMetadata dto.RequestMetadata) *appError.AppError {
	claims, err := s.utilsConfig.JwtService.ValidateToken(ctx, data.Token)
	if err != nil {
		return ErrInvalidToken
	}

	hashPass, err := s.utilsConfig.HashService.HashPassword(data.NewPassword)
	if err != nil {
		return ErrInternal
	}

	pgUUID, appErr := ToPgUUID(claims.UserID)
	if appErr != nil {
		return appErr
	}

	_, err = s.q.UpdateAccount(ctx, db.UpdateAccountParams{
		ID:           pgUUID,
		PasswordHash: hashPass,
	})
	if err != nil {
		return ErrInternal
	}
	return nil
}

func (s *authService) RequestEmailVerification(ctx context.Context, data dto.RequestEmailVerificationRequest, reqMetadata dto.RequestMetadata) *appError.AppError {
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAccountNotFound
		}
		return ErrInternal
	}

	sec, err := s.q.GetAccountSecurityByAccountID(ctx, account.ID)
	if err != nil {
		return ErrInternal
	}

	if sec.IsEmailVerified {
		return ErrEmailAlreadyVerified
	}

	if _, err := s.utilsConfig.JwtService.IssueToken(ctx, FormatUUID(account.ID)); err != nil {
		return ErrInternal
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, data dto.VerifyEmailRequest, reqMetadata dto.RequestMetadata) *appError.AppError {
	claims, err := s.utilsConfig.JwtService.ValidateToken(ctx, data.Token)
	if err != nil {
		return ErrInvalidToken
	}

	pgUUID, appErr := ToPgUUID(claims.UserID)
	if appErr != nil {
		return appErr
	}

	_, err = s.q.UpdateAccountSecurityByAccountID(ctx, db.UpdateAccountSecurityByAccountIDParams{
		AccountID_2:     pgUUID,
		IsEmailVerified: true,
		EmailVerifiedAt: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now(),
		},
	})
	if err != nil {
		return ErrInternal
	}

	return nil
}
