package services

import (
	"context"

	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/dto"
)

type AuthService interface {
	Signin(ctx context.Context, dto dto.SigninRequest, reqMetadata dto.RequestMetadata) (*dto.AuthResponse, *errors.AppError)
	SignUp(ctx context.Context, dto dto.SignupRequest, reqMetadata dto.RequestMetadata) (*dto.AuthResponse, *errors.AppError)
	RequestEmailVerification(ctx context.Context, dto dto.RequestEmailVerificationRequest, reqMetadata dto.RequestMetadata) *errors.AppError
	RequestPasswordReset(ctx context.Context, dto dto.RequestPasswordResetRequest, reqMetadata dto.RequestMetadata) *errors.AppError
	VerifyEmail(ctx context.Context, dto dto.VerifyEmailRequest, reqMetadata dto.RequestMetadata) *errors.AppError
	ResetPassword(ctx context.Context, dto dto.ResetPasswordRequest, reqMetadata dto.RequestMetadata) *errors.AppError
}

type authService struct {
	q           *db.Queries
	utilsConfig *UtilsConfig
	env         *env.Env
}

func NewAuth(queries *db.Queries, cfg *UtilsConfig, env *env.Env) AuthService {
	return &authService{
		q: queries,
		utilsConfig: &UtilsConfig{
			HashService:      cfg.HashService,
			JwtService:       cfg.JwtService,
			SecureJWTService: cfg.SecureJWTService,
			Storage:          cfg.Storage,
		},
		env: env,
	}
}

type UserService interface {
	UpdateProfile(ctx context.Context, dto dto.UpdateProfileRequest, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) (*db.Profile, *errors.AppError)
}

type userService struct {
	q           *db.Queries
	utilsConfig *UtilsConfig
	env         *env.Env
}

func NewUserService(queries *db.Queries, cfg *UtilsConfig, env *env.Env) UserService {
	return &userService{
		q: queries,
		utilsConfig: &UtilsConfig{
			HashService:      cfg.HashService,
			JwtService:       cfg.JwtService,
			SecureJWTService: cfg.SecureJWTService,
			Storage:          cfg.Storage,
		},
		env: env,
	}
}
