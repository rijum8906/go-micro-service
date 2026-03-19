// Package profile contains services for the profile service.
package profile

import (
	"github.com/rijum8906/relay/packages/common/env"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

type ProfileService interface {
	// Core functions
	GetProfile()
	GetProfilesByAccountID()
	UpdateProfile()
	UpdateDisplayName()
	UpdateAvatarURL()
	UpdateName()

	// Public Access
	GetProfileByUsername()
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

type ProfileRepository interface {
	CreateProfile()
	GetProfilesByAccountID()
	GetProfile()
	UpdateProfile()
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
