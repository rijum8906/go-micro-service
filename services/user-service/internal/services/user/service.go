package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreutils"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/protoutils"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func (s *userService) GenerateScopedToken(ctx context.Context, req *userv1.GenerateScopedTokenRequest, user *dto.UserInfo) (*userv1.GenerateScopedTokenResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("generate scoped token request is required")
	}

	if user == nil || user.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	if req.AuthMethod != corev1.AuthMethod_AUTH_METHOD_PASSWORD {
		return nil, apperror.ErrValidation.WithMessage("invalid auth method")
	}

	scopedToken, appErr := s.utils.TokenManager.IssueScopedToken(ctx, user.UserID, token.TokenScope(req.GetScope()))
	if appErr != nil {
		return nil, appErr
	}

	userID, _ := uuid.Parse(user.UserID)

	userInfo, appErr := s.repos.User.GetUser(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	if appErr = s.utils.HashService.Verify(userInfo.PasswordHash.String, req.AuthValue); appErr != nil {
		return nil, appErr
	}

	return &userv1.GenerateScopedTokenResponse{
		Token: &modelsv1.Token{
			Value:     scopedToken,
			ExpiresAt: coreutils.ParseToProtoTimestamp(s.env.ScopedTokenTTL),
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

func (s *userService) UpdateProfileName(ctx context.Context, req *userv1.UpdateProfileNameRequest, userInfo *dto.UserInfo) (*modelsv1.Profile, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update profile name request is required")
	}

	profileID, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	if appErr := s.validateProfileAccess(ctx, profileID, userInfo); appErr != nil {
		return nil, appErr
	}

	profile, appErr := s.repos.Profile.UpdateProfileNames(ctx, profileID, req.GetFirstName(), req.GetLastName())
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(profile), nil
}

func (s *userService) UpdateProfileAvatarUrl(ctx context.Context, req *userv1.UpdateProfileAvatarUrlRequest, userInfo *dto.UserInfo) (*modelsv1.Profile, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update profile avatar request is required")
	}

	profileID, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	if appErr := s.validateProfileAccess(ctx, profileID, userInfo); appErr != nil {
		return nil, appErr
	}

	profile, appErr := s.repos.Profile.UpdateProfileAvatar(ctx, profileID, req.GetAvatarUrl())
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(profile), nil
}

func (s *userService) GetProfile(ctx context.Context, userInfo *dto.UserInfo) (*modelsv1.Profile, *apperror.AppError) {
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	profile, appErr := s.repos.Profile.GetProfileByUserID(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(profile), nil
}

func (s *userService) GetUser(ctx context.Context, userInfo *dto.UserInfo) (*modelsv1.User, *apperror.AppError) {
	if userInfo == nil || userInfo.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return nil, appErr
	}

	user, appErr := s.repos.User.GetUser(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapUser(user), nil
}

func (s *userService) validateProfileAccess(ctx context.Context, profileID uuid.UUID, userInfo *dto.UserInfo) *apperror.AppError {
	if userInfo == nil || userInfo.UserID == "" {
		return apperror.ErrValidation.WithMessage("user metadata is required")
	}

	userID, appErr := utils.NewUUID(userInfo.UserID)
	if appErr != nil {
		return appErr
	}

	profile, appErr := s.repos.Profile.GetProfileByUserID(ctx, userID)
	if appErr != nil {
		return appErr
	}

	if profile.ID != profileID {
		return apperror.ErrForbidden.WithMessage("profile does not belong to user")
	}

	return nil
}

func (s *userService) CheckExists(ctx context.Context, id string) (bool, *apperror.AppError) {
	useID, err := uuid.Parse(id)
	if err != nil {
		return false, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}
	exists, appErr := s.repos.User.CheckExists(ctx, useID)
	if appErr != nil {
		return false, appErr
	}

	return exists, nil
}

func (s *userService) CheckEmailExists(ctx context.Context, email string) (bool, *apperror.AppError) {
	if appErr := protoutils.ValidateEmailReq(&corev1.EmailRequest{Email: email}); appErr != nil {
		return false, appErr
	}

	exists, err := s.repos.User.CheckEmailExists(ctx, email)
	if err != nil {
		return false, apperror.ErrInternal.WithMessage("Failed to check user exists").WithDetail("error", err.Error())
	}

	return exists, nil
}
