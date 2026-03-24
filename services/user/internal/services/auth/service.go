package auth

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/token"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/dto"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func (s *authService) Login(ctx context.Context, data dto.Login, meta *dto.RequestMeta) (*dto.AuthResult, *apperror.AppError) {
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

	session, appErr := s.repos.Session.CreateSession(ctx, db.CreateSessionParams{
		UserID:           user.ID,
		UserAgent:        meta.UserAgent,
		IpAddr:           meta.IPAddr,
		DeviceID:         meta.DeviceID,
		RefreshTokenHash: refreshTokenHash,
	})
	if appErr != nil {
		return nil, appErr
	}

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.ID.String(), session.ID.String(), meta.DeviceID, token.ScopeAuth)
	if appErr != nil {
		return nil, appErr
	}

	return &dto.AuthResult{
		User:    user,
		Profile: profile,
		Tokens:  utils.ParseTokens(accessToken, refreshTokenHash),
	}, nil
}
