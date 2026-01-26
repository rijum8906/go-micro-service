// Package services
package services

import (
	"context"

	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

type AuthenticationService interface {
	Signin(ctx context.Context, dto dto.SignInDTO) (*db.Account, error)
	SignUp(ctx context.Context, dto dto.SignUpDTO) (*db.Account, error)
}

type authenticationService struct {
	q           *db.Queries
	utilsConfig *UtilsConfig
}

func NewAuthenticationService(queries *db.Queries, cfg *UtilsConfig) AuthenticationService {
	return &authenticationService{
		q: queries,
		utilsConfig: &UtilsConfig{
			HashService: cfg.HashService,
			JwtService:  cfg.JwtService,
		},
	}
}

func (s *authenticationService) Signin(ctx context.Context, dto dto.SignInDTO) (*db.Account, error) {
	return nil, nil
}

func (s *authenticationService) SignUp(ctx context.Context, dto dto.SignUpDTO) (*db.Account, error) {
	return nil, nil
}
