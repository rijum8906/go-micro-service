package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	coreconstants "github.com/rijum8906/relay/packages/core/constants"
	"github.com/rijum8906/relay/packages/core/coreutils"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func (s *UserService) GenerateScopedToken(ctx context.Context, req *userv1.GenerateScopedTokenRequest) (*userv1.GenerateScopedTokenResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("revoke other sessions request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	if req.AuthMethod.String() != string(coreconstants.AuthMethodPassword) {
		return nil, apperror.ErrValidation.WithMessage("invalid auth method")
	}
	if !constants.IsValidaTokenScope(req.Scope.String()) {
		return nil, apperror.ErrValidation.WithMessage("invalid token scope")
	}

	userID, _ := uuid.Parse(userInfo.UserID)

	user, err := s.DBQ.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound.WithMessage("user not found")
		}
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to get user by id").
			WithDetail("db_error", err.Error())
	}

	if !s.HashService.Verify(user.PasswordHash.String, req.AuthValue) {
		return nil, apperror.ErrValidation.WithMessage("invalid password")
	}

	tokenRes, appErr := s.TokenManager.GenerateToken(
		userInfo.UserID,
		uuid.NewString(),
		constants.TokenScopeChangePassword,
		s.Config.ScopedTokenTTL,
	)
	if appErr != nil {
		return nil, appErr
	}

	return &userv1.GenerateScopedTokenResponse{
		Token: &modelsv1.Token{
			Value:     tokenRes.TokenString,
			ExpiresAt: coreutils.ParseToProtoTimestamp(s.Config.ScopedTokenTTL),
		},
	}, nil
}

func (s *UserService) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("change password request is required")
	}

	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	claims, appErr := s.TokenManager.ValidateScopedToken(ctx, req.TokenScope)
	if appErr != nil {
		return nil, appErr
	}
	if claims.Subject != userInfo.UserID {
		return nil, apperror.ErrValidation.WithMessage("invalid user id in token")
	}

	if claims.Scope != constants.TokenScopeChangePassword {
		return nil, apperror.ErrValidation.WithMessage("invalid scoped token scope for change password")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	newPasswordHash, appErr := s.HashService.Hash(req.NewPassword)
	if appErr != nil {
		return nil, appErr
	}

	if err = s.DBQ.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID: userID,
		PasswordHash: pgtype.Text{
			String: newPasswordHash,
			Valid:  true,
		},
	}); err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to update user").WithDetail("error", err.Error())
	}

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
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
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("Failed to update profile").
			WithDetail("db_error", err.Error())
	}

	return utils.MapProfile(&profile), nil
}

func (s *UserService) UpdateProfileAvatarUrl(ctx context.Context, req *userv1.UpdateProfileAvatarUrlRequest) (*modelsv1.Profile, error) {
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
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("Failed to update profile").
			WithDetail("db_error", err.Error())
	}

	return utils.MapProfile(&profile), nil
}

func (s *UserService) GetProfile(ctx context.Context, req *corev1.EmptyRequest) (*modelsv1.Profile, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	profile, appErr := s.DBQ.GetProfileByUserID(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(&profile), nil
}

func (s *UserService) GetUser(ctx context.Context, req *corev1.EmptyRequest) (*modelsv1.User, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	user, appErr := s.DBQ.GetUser(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapUser(&user), nil
}

func (s *UserService) CheckEmailExists(ctx context.Context, req *corev1.EmailRequest) (*userv1.CheckExistsResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("check email exists request is required")
	}

	exists, err := s.DBQ.CheckUserEmailExists(ctx, req.Email)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to check user exists").WithDetail("error", err.Error())
	}

	return &userv1.CheckExistsResponse{
		Exists: exists,
	}, nil
}

func (s *UserService) CheckExists(ctx context.Context, req *corev1.IDRequest) (*userv1.CheckExistsResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("check email exists request is required")
	}
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	exists, err := s.DBQ.CheckUserExists(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to check user exists").WithDetail("error", err.Error())
	}

	return &userv1.CheckExistsResponse{
		Exists: exists,
	}, nil
}
