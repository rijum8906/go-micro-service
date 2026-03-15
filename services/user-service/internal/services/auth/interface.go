// Package auth contains services for the auth service.
package auth

import (
	"context"

	"github.com/rijum8906/go-micro-service/packages/common/env"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/api/dto/response"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/utils"
)

type AuthService interface {
	Signin(ctx context.Context, dto request.SigninRequest, reqMetadata request.RequestMetadata) (*response.AuthResponse, *errors.AppError)
	SignUp(ctx context.Context, dto request.SignupRequest, reqMetadata request.RequestMetadata) (*response.AuthResponse, *errors.AppError)

	RequestEmailVerification(ctx context.Context, dto request.RequestEmailVerificationRequest, reqMetadata request.RequestMetadata) *errors.AppError
	RequestPasswordReset(ctx context.Context, dto request.RequestPasswordResetRequest, reqMetadata request.RequestMetadata) *errors.AppError
	VerifyEmail(ctx context.Context, dto request.VerifyEmailRequest, reqMetadata request.RequestMetadata) *errors.AppError
	ResetPassword(ctx context.Context, dto request.ResetPasswordRequest, reqMetadata request.RequestMetadata) *errors.AppError
}

type authService struct {
	q           *db.Queries
	utilsConfig *utils.UtilsConfig
	env         *env.Env
}

func NewAuth(queries *db.Queries, cfg *utils.UtilsConfig, env *env.Env) AuthService {
	return &authService{
		q: queries,
		utilsConfig: &utils.UtilsConfig{
			HashService:      cfg.HashService,
			JwtService:       cfg.JwtService,
			SecureJWTService: cfg.SecureJWTService,
			Storage:          cfg.Storage,
		},
		env: env,
	}
}
