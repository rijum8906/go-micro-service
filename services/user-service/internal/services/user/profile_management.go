package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func (s *UserService) GetProfile(ctx context.Context, req *corev1.EmptyRequest) (*modelsv1.Profile, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		return nil, constants.ErrUserNotFoundInCtx
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, constants.ErrInvalidUserIDInUserInfo
	}

	profile, appErr := s.DBQ.GetProfileByUserID(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(&profile), nil
}

func (s *UserService) UpdateProfileName(ctx context.Context, req *userv1.UpdateProfileNameRequest) (*modelsv1.Profile, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update profile name request is required")
	}

	profileID, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	profile, err := s.DBQ.UpdateProfileName(ctx, db.UpdateProfileNameParams{
		ID:        profileID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if appErr := utils.AssertRowExists(err, "profile", profileID.String()); appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(&profile), nil
}

func (s *UserService) UpdateProfileAvatarURL(ctx context.Context, req *userv1.UpdateProfileAvatarURLRequest) (*modelsv1.Profile, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update profile avatar request is required")
	}

	profileID, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	profile, err := s.DBQ.UpdateProfileAvatarURL(ctx, db.UpdateProfileAvatarURLParams{
		ID:        profileID,
		AvatarUrl: req.AvatarUrl,
	})
	if appErr := utils.AssertRowExists(err, "profile", profileID.String()); appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(&profile), nil
}
