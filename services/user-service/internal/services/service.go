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

type AccountService interface {
	DeleteAccount(ctx context.Context, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) *errors.AppError
	MyAccount(ctx context.Context, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) (*dto.MyAccountResult, *errors.AppError)
}

type accountService struct {
	q           *db.Queries
	utilsConfig *UtilsConfig
	env         *env.Env
}

func NewAccountService(queries *db.Queries, cfg *UtilsConfig, env *env.Env) AccountService {
	return &accountService{
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

type ProfileService interface {
	GetProfile(ctx context.Context, id string) (*dto.GetProfileResult, *errors.AppError)
	UpdateProfile(ctx context.Context, data dto.UpdateProfileRequest, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) (*db.Profile, *errors.AppError)
	CreateProfile(ctx context.Context, data dto.CreateProfileRequest, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) (*db.Profile, *errors.AppError)
	DeleteProfile(ctx context.Context, data dto.DeleteProfileRequest, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) *errors.AppError
	MyProfile(ctx context.Context, reqMetadata dto.RequestMetadata, authzMetadata dto.AuthzMetadata) (*db.Profile, *errors.AppError)
}

type profileService struct {
	q           *db.Queries
	utilsConfig *UtilsConfig
	env         *env.Env
}

func NewProfileService(queries *db.Queries, cfg *UtilsConfig, env *env.Env) ProfileService {
	return &profileService{
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
