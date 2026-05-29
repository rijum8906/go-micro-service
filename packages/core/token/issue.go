package token

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type TokenResponse struct {
	TokenString string
	TokenID     string
	Subject     string
}

const TokenScopeAuth = "USER_AUTH"

// IssueScopedToken generates a one-time JWT for specific operations (e.g., email verification, password change, user invites).
//
// Core:
//   - Creates a JWT with a random UUID as ID, a subject, and a scope.
//   - Stores token ID as key with "active" status in Redis.
//
// Example:
//   - Scope: USER_EMAIL_VERIFY, Subject: user's email
//   - Scope: USER_PASSWORD_CHANGE, Subject: user ID (must be issued via authentication)
func (tm *tokenManager) IssueScopedToken(ctx context.Context, subject, scope string) (*TokenResponse, *apperror.AppError) {
	tokenID := uuid.NewString()
	return tm.issueToken(ctx, tokenID, subject, scope, "", tm.config.ScopedSecret, tm.config.ScopedTokenTTL)
}

// IssueAuthToken generates a session JWT for authenticated users.
//
// Core:
//   - Creates a JWT with a random UUID as ID, userID as subject, and TokenScopeAuth as scope.
//   - Stores token ID as key with "active" status in Redis.
//   - Adds token ID to user's set for easy revocation and cleanup of old tokens.
//
// Use Case:
//   - User sign-in
func (tm *tokenManager) IssueAuthToken(ctx context.Context, userID, sessionID string) (*TokenResponse, *apperror.AppError) {
	tokenID := uuid.NewString()
	userTokenKey := fmt.Sprintf("user_tokens:%s", userID)

	// Issue the token
	tokenRes, appErrr := tm.issueToken(ctx, tokenID, userID, TokenScopeAuth, sessionID, tm.config.JwtSecret, tm.config.SessionTTL)
	if appErrr != nil {
		return nil, appErrr
	}
	// Use Pipeline for ALL Redis operations (SINGLE round trip)
	pipe := tm.redisClient.Pipeline()

	// 1. Store individual token
	pipe.Set(ctx, tokenID, "active", tm.config.SessionTTL)

	// 2. Add token ID to user's set
	pipe.SAdd(ctx, userTokenKey, tokenID)

	// 3. Set expiry on user's token set
	pipe.Expire(ctx, userTokenKey, tm.config.SessionTTL)

	// Execute all 3 commands in ONE round trip
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to store token data").
			WithDetail("redis_error", err.Error())
	}

	return tokenRes, nil
}

// issueToken is the internal helper for token issuance
func (tm *tokenManager) issueToken(ctx context.Context, tokenID, subject, scope, sessionID string, secret []byte, ttl time.Duration) (*TokenResponse, *apperror.AppError) {
	token := generateTokenClaims(subject, tokenID, scope, sessionID, ttl)

	tokenStr, err := token.SignedString(secret)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to sign token").
			WithDetail("redis_error", err.Error())
	}

	// Store individual token key
	if err = tm.redisClient.Set(ctx, tokenID, "active", ttl).Err(); err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to store token").
			WithDetail("error", err.Error())
	}

	return &TokenResponse{TokenString: tokenStr, TokenID: tokenID, Subject: subject}, nil
}
