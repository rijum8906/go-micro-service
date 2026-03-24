// Package utils
package utils

import (
	"github.com/rijum8906/relay/packages/core/hash"
	"github.com/rijum8906/relay/packages/core/token"
	"github.com/rijum8906/relay/services/user/internal/repository/profile"
	"github.com/rijum8906/relay/services/user/internal/repository/session"
	"github.com/rijum8906/relay/services/user/internal/repository/user"
)

type Utils struct {
	TokenManager *token.TokenManager
	HashService  hash.HashService
}

type Repos struct {
	User    user.UserRepository
	Profile profile.ProfileRepository
	Session session.SessionRepository
}

func NewUtils(tokenManager *token.TokenManager, hashService hash.HashService) *Utils {
	return &Utils{
		TokenManager: tokenManager,
		HashService:  hashService,
	}
}
