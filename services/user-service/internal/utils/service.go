// Package utils
package utils

import (
	"net/url"
	pathPkg "path"

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

func NewTokenURL(token, baseURL, path string) (string, *apperror.AppError) {
	// Basic presence checks
	if token == "" || baseURL == "" || path == "" {
		return "", apperror.ErrValidation.WithMessage("token, base url, and path are required")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", apperror.ErrInternal.WithMessage("failed to parse url").WithDetail("error", err.Error())
	}

	if base.Scheme == "" || base.Host == "" {
		return "", apperror.ErrValidation.
			WithMessage("base url must be an absolute URL (e.g., https://example.com)").
			WithDetail("provided", baseURL)
	}

	// Safely join path
	base.Path = pathPkg.Join(base.Path, path)

	// Encode query
	q := base.Query()
	q.Set("token", token)
	base.RawQuery = q.Encode()

	return base.String(), nil
}
