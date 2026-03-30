package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user/models/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/dto"
	"github.com/rijum8906/relay/services/user/internal/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *authService) Login(ctx context.Context, data dto.Login, client *metadata.ClientInfo) (*authv1.AuthResponse, *apperror.AppError) {
	user, appErr := s.repos.User.GetUserByEmail(ctx, data.Email)
	if appErr != nil {
		return nil, appErr
	}

	appErr = s.utils.HashService.Verify(user.PasswordHash.String, data.Password)
	if appErr != nil {
		return nil, appErr
	}

	refreshTokenHash, appErr := s.utils.HashService.Generate(32)
	if appErr != nil {
		return nil, appErr
	}

	profile, appErr := s.repos.Profile.GetProfileByUserID(ctx, user.ID)
	if appErr != nil {
		return nil, appErr
	}

	timeStamp := pgtype.Timestamptz{
		Time:  time.Now().Add(s.env.SessionTTL),
		Valid: true,
	}
	session, appErr := s.repos.Session.CreateSession(ctx, db.CreateSessionParams{
		UserID:           user.ID,
		UserAgent:        client.UserAgent,
		IpAddr:           client.IPAddress,
		DeviceID:         client.DeviceID,
		ExpiresAt:        timeStamp,
		RefreshTokenHash: refreshTokenHash,
	})
	if appErr != nil {
		return nil, appErr
	}

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.ID.String(), session.ID.String(), client.DeviceID, token.TokenScopeAuth)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

func (s *authService) Register(ctx context.Context, data dto.Register, client *metadata.ClientInfo) (*authv1.AuthResponse, *apperror.AppError) {
	_, appErr := s.repos.User.GetUserByEmail(ctx, data.Email)
	if appErr != nil && appErr.Type != apperror.TypeNotFound {
		return nil, appErr
	}

	hashedPass, appErr := s.utils.HashService.Hash(data.Password)
	if appErr != nil {
		return nil, appErr
	}
	data.Password = hashedPass

	user, appErr := s.repos.User.CreateUser(ctx, &dto.Register{
		Email:     data.Email,
		Password:  data.Password,
		FirstName: data.FirstName,
		LastName:  data.LastName,
	})
	if appErr != nil {
		return nil, appErr
	}

	profile, appErr := s.repos.Profile.CreateProfile(ctx, db.CreateProfileParams{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		UserID:    user.ID,
	})
	if appErr != nil {
		return nil, appErr
	}

	refreshTokenHash, appErr := s.utils.HashService.Generate(32)
	if appErr != nil {
		return nil, appErr
	}

	timeStamp := pgtype.Timestamptz{
		Time:  time.Now().Add(s.env.SessionTTL),
		Valid: true,
	}
	session, appErr := s.repos.Session.CreateSession(ctx, db.CreateSessionParams{
		UserID:           user.ID,
		UserAgent:        client.UserAgent,
		IpAddr:           client.IPAddress,
		DeviceID:         client.DeviceID,
		ExpiresAt:        timeStamp,
		RefreshTokenHash: refreshTokenHash,
	})
	if appErr != nil {
		return nil, appErr
	}

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.ID.String(), session.ID.String(), client.DeviceID, token.TokenScopeAuth)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

func (s *authService) Logout(ctx context.Context, client *metadata.UserInfo) (bool, *apperror.AppError) {
	sessionID, err := uuid.Parse(client.SessionID)
	if err != nil {
		return false, apperror.ErrInternal.WithMessage("Failed to parse session ID").WithDetail("error", err.Error())
	}

	appErr := s.repos.Session.RevokeSession(ctx, sessionID)
	if appErr != nil {
		return false, appErr
	}

	return true, nil
}

func (s *authService) GenerateScopedToken(ctx context.Context, data dto.GenerateScopedToken, user *metadata.UserInfo) (*authv1.ScopedTokenResponse, *apperror.AppError) {
	if user == nil || user.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	scopedToken, appErr := s.utils.TokenManager.IssueScopedToken(ctx, user.UserID, token.TokenScope(data.Scope))
	if appErr != nil {
		return nil, appErr
	}

	return &authv1.ScopedTokenResponse{
		Token: &modelsv1.Token{
			Value:     scopedToken,
			ExpiresIn: timestamppb.New(time.Now().Add(s.env.ScopedTokenTTL)),
		},
	}, nil
}

func (s *authService) ChangePassword(ctx context.Context, data dto.ChangePassword, user *metadata.UserInfo) (*authv1.MutationResponse, *apperror.AppError) {
	if user == nil || user.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	dbUser, appErr := s.repos.User.GetUser(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	appErr = s.utils.HashService.Verify(dbUser.PasswordHash.String, data.CurrentPassword)
	if appErr != nil {
		return nil, appErr
	}

	newPasswordHash, appErr := s.utils.HashService.Hash(data.NewPassword)
	if appErr != nil {
		return nil, appErr
	}

	appErr = s.repos.User.UpdateUserPassword(ctx, userID, newPasswordHash)
	if appErr != nil {
		return nil, appErr
	}

	return &authv1.MutationResponse{
		Success: true,
		Message: "password changed successfully",
	}, nil
}

func (s *authService) UpdateProfileName(ctx context.Context, data dto.UpdateProfileName) (*modelsv1.Profile, *apperror.AppError) {
	profileID, err := uuid.Parse(data.ProfileID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	profile, appErr := s.repos.Profile.UpdateProfileNames(ctx, profileID, data.FirstName, data.LastName)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(profile), nil
}

func (s *authService) UpdateProfileAvatarUrl(ctx context.Context, data dto.UpdateProfileAvatarUrl) (*modelsv1.Profile, *apperror.AppError) {
	profileID, err := uuid.Parse(data.ProfileID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	profile, appErr := s.repos.Profile.UpdateProfileAvatar(ctx, profileID, data.AvatarURL)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(profile), nil
}
