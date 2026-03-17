// Package profile contains services for the profile service.
package profile

import (
	"context"

	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/errors"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/response"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

type ProfileService interface {
	GetProfile(ctx context.Context, id string) (*response.GetProfileResult, *errors.AppError)
	UpdateProfile(ctx context.Context, data request.UpdateProfileRequest, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*db.Profile, *errors.AppError)
	CreateProfile(ctx context.Context, data request.CreateProfileRequest, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*db.Profile, *errors.AppError)
	DeleteProfile(ctx context.Context, id string, authzMetadata request.AuthzMetadata) *errors.AppError
	MyProfile(ctx context.Context, reqMetadata request.RequestMetadata, authzMetadata request.AuthzMetadata) (*db.Profile, *errors.AppError)
}

type profileService struct {
	q           *db.Queries
	utilsConfig *utils.UtilsConfig
	env         *env.Env
}

func NewProfileService(queries *db.Queries, cfg *utils.UtilsConfig, env *env.Env) ProfileService {
	return &profileService{
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
