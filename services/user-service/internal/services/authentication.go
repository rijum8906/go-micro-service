// Package services
package services

import (
	"context"
	"database/sql"
	"errors"

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
	accessToken, err := s.utilsConfig.JwtService.IssueToken(ctx, FormatUUID(account.ID))
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

func (s *authService) SignUp(ctx context.Context, data dto.SignUpDTO) (*dto.AuthenticationResult, error) {
	_, err := s.q.GetAccountByEmail(ctx, data.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// TODO: Handle the error
			return nil, err
		}
	}

	passHash, err := s.utilsConfig.HashService.HashPassword(data.Password)
	if err != nil {
		// TODO: Handle the error
		return nil, err
	}
	createAccountParams := db.CreateAccountParams{
		Email:        data.Email,
		PasswordHash: passHash,
	}
	creeateProfileParams := db.CreateProfileParams{
		FirstName: data.FirstName,
		LastName:  data.LastName,
	}

	account, err := s.q.CreateAccount(ctx, createAccountParams)
	if err != nil {
		// TODO: Handle the error
		return nil, err
	}

	profile, err := s.q.CreateProfile(ctx, creeateProfileParams)
	if err != nil {
		// TODO: Handle the error
		return nil, err
	}

	refreshToken, err := s.utilsConfig.HashService.GenerateRefreshToken()
	if err != nil {
		// TODO: Handle the error
		return nil, err
	}
	accessToken, err := s.utilsConfig.JwtService.IssueToken(ctx, FormatUUID(account.ID))
	if err != nil {
		// TODO: Handle the error
		return nil, err
	}

	return &dto.AuthenticationResult{
		Account:  &account,
		Profiles: []*db.Profile{&profile},
		Tokens: &dto.Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}
