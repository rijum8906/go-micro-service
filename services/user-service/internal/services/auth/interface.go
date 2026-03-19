// Package auth contains services for the auth service.
package auth

import (
	"context"

	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/errors"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

type AuthService interface {
	// Core
	Signin(ctx context.Context, req *user_servicev1.SigninRequest) (*user_servicev1.AuthResponse, *errors.AppError)
	Signup(ctx context.Context, req *user_servicev1.SignupRequest) (*user_servicev1.AuthResponse, *errors.AppError)
	Logout(ctx context.Context, req *user_servicev1.SignoutRequest, auth request.AuthzMetadata) (*user_servicev1.SignoutResponse, *errors.AppError)
	LogoutAllDevices(ctx context.Context, req *user_servicev1.SignoutAllRequest, auth request.AuthzMetadata) (*user_servicev1.SignoutAllResponse, *errors.AppError)

	// Email Verification
	RequestEmailVerification(ctx context.Context, req *user_servicev1.RequestEmailVerificationRequest) *errors.AppError
	VerifyEmail(ctx context.Context, req *user_servicev1.VerifyEmailRequest) *errors.AppError

	// Password Management
	RequestPasswordReset(ctx context.Context, req *user_servicev1.RequestPasswordResetRequest) *errors.AppError
	ResetPassword(ctx context.Context, req *user_servicev1.ResetPasswordRequest) *errors.AppError
	ChangePassword(ctx context.Context, req *user_servicev1.ChangePasswordRequest, authzMetadata request.AuthzMetadata) *errors.AppError

	// Session Management
	GetSessions(ctx context.Context, req *user_servicev1.GetSessionsRequest, authzMetadata request.AuthzMetadata) (*user_servicev1.GetSessionsResponse, *errors.AppError)
	RevokeSession(ctx context.Context, req *user_servicev1.RevokeSessionRequest, authzMetadata request.AuthzMetadata) (*user_servicev1.RevokeSessionResponse, *errors.AppError)
	RevokeAllSessions(ctx context.Context, req *user_servicev1.RevokeAllSessionsRequest, authzMetadata request.AuthzMetadata) (*user_servicev1.RevokeAllSessionsResponse, *errors.AppError)
}

type authService struct {
	q           *db.Queries
	utilsConfig *utils.UtilsConfig
	env         *env.Env
}

func NewAuthService(queries *db.Queries, cfg *utils.UtilsConfig, env *env.Env) AuthService {
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

type AuthRepository interface {
	CreateSession()
	GetSession()
	RevokeSession()
	RevokeAllSessions()
	DeleteExpiredSessions()
}

type authRepository struct {
	q           *db.Queries
	utilsConfig *utils.UtilsConfig
	env         *env.Env
}

func NewAuthRepository(queries *db.Queries, cfg *utils.UtilsConfig, env *env.Env) AuthRepository {
	return &authRepository{
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
