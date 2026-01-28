package services

import (
	"context"

	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

type AuthService interface {
	Signin(ctx context.Context, dto dto.SignInDTO) (*dto.AuthenticationResult, error)
	SignUp(ctx context.Context, dto dto.SignUpDTO) (*dto.AuthenticationResult, error)
}

type authService struct {
	q           *db.Queries
	utilsConfig *UtilsConfig
}

func NewAuth(queries *db.Queries, cfg *UtilsConfig) AuthService {
	return &authService{
		q: queries,
		utilsConfig: &UtilsConfig{
			HashService: cfg.HashService,
			JwtService:  cfg.JwtService,
		},
	}
}
