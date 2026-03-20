// Package profile contains services for the profile service.
package profile

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	profilev1 "github.com/rijum8906/relay/packages/pb/user_service/profile/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

type ProfileService interface {
	// Core functions
	CreateProfile(ctx context.Context, data *authv1.SignupRequest, authzMetadata *request.AuthzMetadata) (*db.Profile, *errors.AppError)
	GetProfile(ctx context.Context, profileID pgtype.UUID) (*db.Profile, *errors.AppError)
	GetProfilesByAccountID(ctx context.Context, accountID pgtype.UUID) (*[]db.Profile, *errors.AppError)
	UpdateProfile(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateProfileRequest) (*db.Profile, *errors.AppError)
	UpdateDisplayName(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateDisplayNameRequest) (*db.Profile, *errors.AppError)
	UpdateAvatarURL(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateAvatarUrlRequest) (*db.Profile, *errors.AppError)
	UpdateName(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateNameRequest) (*db.Profile, *errors.AppError)
}

type profileService struct {
	repo        ProfileRepository
	q           *db.Queries
	utilsConfig *utils.UtilsConfig
	env         *env.Env
}

func NewProfileService(repo ProfileRepository, queries *db.Queries, env *env.Env) ProfileService {
	return &profileService{
		repo: repo,
		q:    queries,
		env:  env,
	}
}

type ProfileRepository interface {
	CreateProfile(ctx context.Context, accountID pgtype.UUID, data *authv1.SignupRequest) (*db.Profile, *errors.AppError)
	GetProfilesByAccountID(ctx context.Context, accountID pgtype.UUID) (*[]db.Profile, *errors.AppError)
	GetProfile(ctx context.Context, profileID pgtype.UUID) (*db.Profile, *errors.AppError)
	UpdateProfile(ctx context.Context, profileID pgtype.UUID, data *profilev1.UpdateProfileRequest) (*db.Profile, *errors.AppError)
}

type profileRepository struct {
	q           *db.Queries
	utilsConfig *utils.UtilsConfig
	env         *env.Env
}

func NewProfileRepository(queries *db.Queries, cfg *utils.UtilsConfig, env *env.Env) ProfileRepository {
	return &profileRepository{
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
