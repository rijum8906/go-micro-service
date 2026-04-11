// Package utils
package utils

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/hash"
	"github.com/rijum8906/relay/packages/core/token"
	"github.com/rijum8906/relay/services/user/internal/repository/profile"
	"github.com/rijum8906/relay/services/user/internal/repository/session"
	"github.com/rijum8906/relay/services/user/internal/repository/user"
)

type ServiceUtils struct {
	TokenManager *token.TokenManager
	HashService  hash.HashService
}

type Repos struct {
	User    user.UserRepository
	Profile profile.ProfileRepository
	Session session.SessionRepository
}

func NewUtils(tokenManager *token.TokenManager, hashService hash.HashService) *ServiceUtils {
	return &ServiceUtils{
		TokenManager: tokenManager,
		HashService:  hashService,
	}
}

func NewUUID(id string) (uuid.UUID, *apperror.AppError) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, apperror.ErrValidation.WithMessage("invalid uuid").WithDetail("error", err.Error())
	}

	return u, nil
}
