// Package auth contains services for the auth service.
package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

type AuthService interface {
	// Core
	Signin(ctx context.Context, req *authv1.SigninRequest) (*user_servicev1.AuthenticationResult, *errors.AppError)
	Signup(ctx context.Context, req *authv1.SignupRequest) (*user_servicev1.AuthenticationResult, *errors.AppError)
	Logout(ctx context.Context, req *authv1.LogoutRequest, auth request.AuthzMetadata) (*authv1.LogoutResponse, *errors.AppError)
	LogoutAllDevices(ctx context.Context, req *authv1.LogoutAllDeviceRequest, auth request.AuthzMetadata) (*authv1.LogoutAllDeviceResponse, *errors.AppError)

	// Email Verification
	RequestEmailVerification(ctx context.Context, req *authv1.RequestEmailVerificationRequest) *errors.AppError
	VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) *errors.AppError

	// Password Management
	RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) *errors.AppError
	ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) *errors.AppError
	ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest, authzMetadata request.AuthzMetadata) *errors.AppError

	// Session Management
	GetSessions(ctx context.Context, req *authv1.GetSessionsRequest, authzMetadata request.AuthzMetadata) (*authv1.GetSessionsResponse, *errors.AppError)
	RevokeSession(ctx context.Context, req *authv1.RevokeSessionRequest, authzMetadata request.AuthzMetadata) (*authv1.RevokeSessionResponse, *errors.AppError)
	RevokeAllSessions(ctx context.Context, req *authv1.RevokeAllSessionsRequest, authzMetadata request.AuthzMetadata) (*authv1.RevokeAllSessionsResponse, *errors.AppError)
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
	CreateSession(ctx context.Context, metadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*db.Session, *errors.AppError)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string, authzMetadata request.AuthzMetadata) (*db.Session, *errors.AppError)
	RevokeSession(ctx context.Context, id pgtype.UUID, authzMetadata request.AuthzMetadata) *errors.AppError
	RevokeAllSessions(ctx context.Context, authzMetadata request.AuthzMetadata) *errors.AppError
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
