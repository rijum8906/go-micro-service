// Package services
package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	appError "github.com/rijum8906/go-micro-service/packages/common/errors"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

func ToPgUUID(idStr string) (pgtype.UUID, error) {
	// 1. Parse string into a [16]byte array
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		return pgtype.UUID{}, err
	}

	// 2. Assign the bytes directly to the pgtype structure
	return pgtype.UUID{
		Bytes: parsed,
		Valid: true,
	}, nil
}

func (s *authService) Signin(ctx context.Context, data dto.SigninRequest, reqMetadata dto.RequestMetadata) (*dto.AuthResponse, error) {
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appError.ErrAccountNotFound
		}
		return nil, appError.ErrInternal
	}

	if err := s.utilsConfig.HashService.VerifyPassword(account.PasswordHash, data.Password); err != nil {
		return nil, appError.ErrInvalidCredentials
	}

	profiles, err := s.q.GetProfilesByAccountID(ctx, account.ID)
	if err != nil {
		return nil, appError.ErrInternal
	}

	refreshToken, err := s.utilsConfig.HashService.GenerateRefreshToken()
	if err != nil {
		return nil, appError.ErrInternal
	}

	accessToken, err := s.utilsConfig.JwtService.IssueToken(ctx, FormatUUID(account.ID))
	if err != nil {
		return nil, appError.ErrInternal
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

func (s *authService) SignUp(ctx context.Context, data dto.SignupRequest, reqMetadata dto.RequestMetadata) (*dto.AuthResponse, error) {
	_, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err == nil {
		return nil, appError.ErrAccountExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, appError.ErrInternal
	}

	passHash, err := s.utilsConfig.HashService.HashPassword(data.Password)
	if err != nil {
		return nil, appError.ErrInternal
	}

	account, err := s.q.CreateAccount(ctx, db.CreateAccountParams{
		Email:        data.Email,
		PasswordHash: passHash,
	})
	if err != nil {
		return nil, appError.ErrInternal
	}

	profile, err := s.q.CreateProfile(ctx, db.CreateProfileParams{
		FirstName: data.FirstName,
		LastName:  data.LastName,
	})
	if err != nil {
		return nil, appError.ErrInternal
	}

	refreshToken, err := s.utilsConfig.HashService.GenerateRefreshToken()
	if err != nil {
		return nil, appError.ErrInternal
	}

	accessToken, err := s.utilsConfig.JwtService.IssueToken(ctx, FormatUUID(account.ID))
	if err != nil {
		return nil, appError.ErrInternal
	}

	return &dto.AuthResponse{
		Account:  &account,
		Profiles: &[]db.Profile{profile},
		Token: &dto.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, data dto.RequestPasswordResetRequest, reqMetadata dto.RequestMetadata) error {
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appError.ErrAccountNotFound
		}
		return appError.ErrInternal
	}

	_, err = s.utilsConfig.JwtService.IssueToken(ctx, account.ID.String())
	if err != nil {
		return appError.ErrInternal
	}

	// TODO: Send reset password email

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, data dto.ResetPasswordRequest, reqMetadata dto.RequestMetadata) error {
	claims, err := s.utilsConfig.JwtService.ValidateToken(ctx, data.Token)
	if err != nil {
		return appError.ErrInvalidToken
	}

	accID := claims.UserID
	if accID == "" {
		return appError.ErrInvalidTokenClaims
	}
	hashPass, err := s.utilsConfig.HashService.HashPassword(data.NewPassword)
	if err != nil {
		return appError.ErrInternal
	}
	pgUUID, err := ToPgUUID(accID)
	if err != nil {
		return appError.ErrInvalidTokenClaims
	}
	_, err = s.q.UpdateAccount(ctx, db.UpdateAccountParams{
		ID:           pgUUID,
		PasswordHash: hashPass,
	})
	if err != nil {
		return appError.ErrInternal
	}
	return nil
}

func (s *authService) RequestEmailVerification(ctx context.Context, data dto.RequestEmailVerificationRequest, reqMetadata dto.RequestMetadata) error {
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appError.ErrAccountNotFound
		}
		return appError.ErrInternal
	}

	sec, err := s.q.GetAccountSecurityByAccountID(ctx, account.ID)
	if err != nil {
		return appError.ErrInternal
	}

	if sec.IsEmailVerified {
		return errors.New("email is already verified")
	}

	if _, err := s.utilsConfig.JwtService.IssueToken(ctx, account.ID.String()); err != nil {
		return appError.ErrInternal
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, data dto.VerifyEmailRequest, reqMetadata dto.RequestMetadata) error {
	claims, err := s.utilsConfig.JwtService.ValidateToken(ctx, data.Token)
	if err != nil {
		return appError.ErrInvalidToken
	}
	accID := claims.UserID
	if accID == "" {
		return appError.ErrInvalidTokenClaims
	}
	pgUUID, err := ToPgUUID(accID)
	if err != nil {
		return appError.ErrInvalidTokenClaims
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
		return appError.ErrInternal
	}

	return nil
}
