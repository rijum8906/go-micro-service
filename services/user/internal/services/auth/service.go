package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/dto"
	"github.com/rijum8906/relay/services/user/internal/utils"
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

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.ID.String(), session.ID.String(), client.DeviceID, token.ScopeAuth)
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

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.ID.String(), session.ID.String(), client.DeviceID, token.ScopeAuth)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

func (s *authService) Logout(ctx context.Context, client *metadata.UserInfo) (bool, *apperror.AppError) {
	session, appErr := s.repos.Session.GetSessionByRefreshToken(ctx, client.RefreshToken)
	if appErr != nil {
		return false, appErr
	}

	err := s.repos.Session.RevokeSession(ctx, session.ID)
	if err != nil {
		return false, apperror.ErrInternal.WithMessage("Failed to logout").WithDetail("error", err.Error())
	}

	return false, nil
}
