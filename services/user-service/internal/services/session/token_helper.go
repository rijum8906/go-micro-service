package session

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
)

type TokenPair struct {
	AccessToken      string
	RefreshTokenHash string
}

func (s *SessionService) issueTokenPair(ctx context.Context, userID, sessionID string) (TokenPair, *apperror.AppError) {
	accessTokenRes, appErr := s.TokenManager.IssueAuthToken(ctx, userID, sessionID)
	if appErr != nil {
		return TokenPair{}, apperror.ErrInternal.
			WithDetail("internal_message", "failed to issue access token").
			WithDetail("token_error", appErr.Error())
	}

	refreshTokenHash, appErr := s.HashService.Generate(32)
	if appErr != nil {
		return TokenPair{}, appErr
	}

	return TokenPair{
		AccessToken:      accessTokenRes.TokenString,
		RefreshTokenHash: refreshTokenHash,
	}, nil
}
