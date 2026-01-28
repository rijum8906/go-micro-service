// Package services
package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

func (s *authService) Signin(ctx context.Context, data dto.SignInDTO) (*dto.AuthenticationResult, error) {
	account, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// TODO: Handle the error
			return nil, err
		}
		// TODO: Handle the error
		return nil, err
	}

	err = s.utilsConfig.HashService.VerifyPassword(account.PasswordHash, data.Password)
	if err != nil {
		// TODO: Handle the error
		return nil, err
	}

	profiles, err := s.q.GetProfilesByAccountID(ctx, account.ID)
	if err != nil {
		// TODO: Handle the error
		return nil, err
	}

	refreshToken, err := s.utilsConfig.HashService.GenerateRefreshToken()
	if err != nil {
		// TODO: Handle the error
		return nil, err
	}
	accessToken, err := s.utilsConfig.JwtService.IssueToken(ctx, formatUUID(account.ID))
	if err != nil {
		// TODO: Handle the error
		return nil, err
	}

	return &dto.AuthenticationResult{
		Account:  &account,
		Profiles: make([]*db.Profile, len(profiles)),
		Tokens: &dto.Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

func (s *authService) SignUp(ctx context.Context, dto dto.SignUpDTO) (*dto.AuthenticationResult, error) {
	return nil, nil
}

func formatUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// Formats the 16-byte array into the standard UUID string format
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}
