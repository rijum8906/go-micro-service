// Package auth
package auth

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/dto"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type AuthService interface {
	Login(ctx context.Context, data dto.Login, meta *dto.RequestMeta) (*authv1.AuthResponse, *apperror.AppError)
	Register(ctx context.Context, data dto.Register, meta *dto.RequestMeta) (*authv1.AuthResponse, *apperror.AppError)
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
