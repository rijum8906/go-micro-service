// Package auth
package auth

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/internal/dto"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type AuthService interface {
	Login(ctx context.Context, data dto.Login, meta *dto.RequestMeta) (*dto.AuthResult, *apperror.AppError)
}

type authService struct {
	repos *utils.Repos
	utils *utils.Utils
}

func NewAuthService(repo *utils.Repos, utils *utils.Utils) AuthService {
	return &authService{
		repos: repo,
		utils: utils,
	}
}
