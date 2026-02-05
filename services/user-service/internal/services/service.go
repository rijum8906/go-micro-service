package services

import (
	"context"

	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

type AuthService interface {
	Signin(ctx context.Context, dto dto.SigninRequest, reqMetadata dto.RequestMetadata) (*dto.AuthenticationResult, error)
	SignUp(ctx context.Context, dto dto.SignupRequest, reqMetadata dto.RequestMetadata) (*dto.AuthenticationResult, error)
	RequestEmailVerification(ctx context.Context, dto dto.RequestEmailVerificationRequest, reqMetadata dto.RequestMetadata) error
	RequestPasswordReset(ctx context.Context, dto dto.RequestPasswordResetRequest, reqMetadata dto.RequestMetadata) error
	VerifyEmail(ctx context.Context, dto dto.VerifyEmailRequest, reqMetadata dto.RequestMetadata) error
	ResetPassword(ctx context.Context, dto dto.ResetPasswordRequest, reqMetadata dto.RequestMetadata) error
}

type authService struct {
	q           *db.Queries
	utilsConfig *UtilsConfig
}

func NewAuth(queries *db.Queries, cfg *UtilsConfig) AuthService {
	return &authService{
		q: queries,
		utilsConfig: &UtilsConfig{
			HashService:      cfg.HashService,
			JwtService:       cfg.JwtService,
			SecureJWTService: cfg.SecureJWTService,
		},
	}
}
