package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *userService) GenerateScopedToken(ctx context.Context, req *userv1.GenerateScopedTokenRequest, user *metadata.UserInfo) (*userv1.GenerateScopedTokenResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("generate scoped token request is required")
	}

	if user == nil || user.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	scopedToken, appErr := s.utils.TokenManager.IssueScopedToken(ctx, user.UserID, token.TokenScope(req.GetScope()))
	if appErr != nil {
		return nil, appErr
	}

	return &userv1.GenerateScopedTokenResponse{
		Token: &modelsv1.Token{
			Value:     scopedToken,
			ExpiresIn: timestamppb.New(time.Now().Add(s.env.ScopedTokenTTL)),
		},
	}, nil
}

func (s *userService) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("change password request is required")
	}

	scopedToken := req.GetScopedToken()
	if scopedToken == nil || scopedToken.GetValue() == "" {
		return nil, apperror.ErrValidation.WithMessage("change password scoped token is required")
	}

	claims, appErr := s.utils.TokenManager.ValidateScopedToken(ctx, scopedToken.GetValue())
	if appErr != nil {
		return nil, appErr
	}

	if claims.Scope != token.TokenScopeChangePassword {
		return nil, apperror.ErrValidation.WithMessage("invalid scoped token scope for change password")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	newPasswordHash, appErr := s.utils.HashService.Hash(req.GetNewPassword())
	if appErr != nil {
		return nil, appErr
	}

	appErr = s.repos.User.UpdateUserPassword(ctx, userID, newPasswordHash)
	if appErr != nil {
		return nil, appErr
	}

	if appErr = s.utils.TokenManager.RevokeScopedToken(ctx, scopedToken.GetValue()); appErr != nil {
		return nil, appErr
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *userService) UpdateProfileName(ctx context.Context, req *userv1.UpdateProfileNameRequest) (*modelsv1.Profile, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update profile name request is required")
	}

	profileID, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	profile, appErr := s.repos.Profile.UpdateProfileNames(ctx, profileID, req.GetFirstName(), req.GetLastName())
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(profile), nil
}

func (s *userService) UpdateProfileAvatarUrl(ctx context.Context, req *userv1.UpdateProfileAvatarUrlRequest) (*modelsv1.Profile, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update profile avatar request is required")
	}

	profileID, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	profile, appErr := s.repos.Profile.UpdateProfileAvatar(ctx, profileID, req.GetAvatarUrl())
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(profile), nil
}
