package token

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
)

// ValidateAuthToken validates authentication tokens used for user sessions.
//
// Auth tokens are long-lived session tokens stored in Redis by claims.ID (userID:uuid).
// This allows for:
//   - Efficient token lookup by ID
//   - Centralized revocation control
//   - Pattern-based revocation (userID:*)
//
// The token is valid if:
//  1. JWT signature is valid and not expired (validated by jwt.ParseWithClaims)
//  2. Token scope matches TokenScopeAuth
//  3. Token exists in Redis (not revoked)
//
// Returns the token claims if valid, otherwise returns an AppError.
// NOTE: we need to match the subject with userID to validate the token
func (tm *tokenManager) ValidateAuthToken(ctx context.Context, tokenStr string) (*Claims, *apperror.AppError) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return tm.config.JwtSecret, nil
	})
	if err != nil {
		return nil, MapTokenError(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, apperror.ErrValidation.WithMessage("invalid token claims")
	}

	// Verify this is an auth token, not a scoped token
	if claims.Scope != TokenScopeAuth {
		return nil, apperror.New(apperror.CodeTokenInvalid, "invalid token scope: expected auth token")
	}

	// Check if token exists in Redis (not revoked)
	if err = tm.redisClient.Get(ctx, claims.ID).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperror.New(apperror.CodeTokenExpired, "token has been revoked or never issued")
		}
		return nil, mapRedisError(err)
	}

	return claims, nil
}

// ValidateScopedToken validates one-time or scope-limited tokens (e.g., email verification, invite links, password reset).
//
// Scoped tokens are single-use tokens stored in Redis by claims.ID (UUID).
// This approach ensures:
//   - Each token can only be used once (atomic read-and-delete prevents race conditions)
//   - Tokens are self-contained with all necessary claims in the JWT
//   - Redis acts as a blacklist for consumed tokens rather than storing valid tokens
//
// The token is valid if:
//  1. JWT signature is valid and not expired (validated by jwt.ParseWithClaims)
//  2. Token exists in Redis (hasn't been consumed yet)
//
// IMPORTANT: This method atomically reads AND deletes the token from Redis,
// making it safe for concurrent requests. The first request succeeds,
// subsequent requests with the same token will fail as the token is already consumed.
//
// Returns the token claims if valid, otherwise returns an AppError.
func (tm *tokenManager) ValidateScopedToken(ctx context.Context, tokenStr string) (*Claims, *apperror.AppError) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return tm.config.ScopedSecret, nil
	})
	if err != nil {
		return nil, MapTokenError(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to parse token claims")
	}

	// Your IssueScopedToken stores by UUID, so look up by UUID
	val, err := tm.redisClient.GetDel(ctx, claims.ID).Result()
	if errors.Is(err, redis.Nil) {
		return nil, apperror.New(apperror.CodeTokenExpired, "token has already been used or has expired")
	}
	if err != nil {
		return nil, mapRedisError(err)
	}

	// Verify the stored value matches expected format
	if val != "active" {
		return nil, apperror.New(apperror.CodeTokenInvalid, "invalid token state")
	}

	return claims, nil
}
