package token

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreutils"
)

// IsValidTokenScope validates that a token scope follows the required format.
//
// Rules:
//  1. Must contain only uppercase letters, numbers, and underscores
//  2. Must follow the structure: [ACTION]_[RESOURCE]_[ACTION_NAME]
//  3. Must have at least 2 underscores (3 parts minimum)
//  4. Cannot start or end with underscore
//  5. Cannot have consecutive underscores
//  6. Each part must be non-empty
//
// Examples of valid scopes:
//   - READ_USER_PROFILE
//   - WRITE_POST_CONTENT
//   - DELETE_COMMENT_PERMANENT
//   - USER_AUTH (simple fallback)
//   - SEND_EMAIL_VERIFICATION
//   - MANAGE_ADMIN_SETTINGS
//
// Examples of invalid scopes:
//   - read_user_profile (lowercase)
//   - READ_USER (only 2 parts)
//   - READ_USER_ (trailing underscore)
//   - _READ_USER (leading underscore)
//   - READ__USER (double underscore)
//   - READ_USER_123 (numbers allowed but structure matters)
//
// Returns true if the scope is valid, false otherwise.
func IsValidTokenScope(scope string) bool {
	// Must not be empty
	if scope == "" {
		return false
	}

	// Must contain only uppercase letters, numbers, and underscores
	matched, err := regexp.MatchString(`^[A-Z0-9_]+$`, scope)
	if err != nil || !matched {
		return false
	}

	// Cannot start or end with underscore
	if strings.HasPrefix(scope, "_") || strings.HasSuffix(scope, "_") {
		return false
	}

	// Cannot have consecutive underscores
	if strings.Contains(scope, "__") {
		return false
	}

	// Must have at least 2 underscores (3 parts)
	parts := strings.Split(scope, "_")
	if len(parts) < 3 {
		return false
	}

	// Each part must be non-empty (already guaranteed by previous checks)
	return !slices.Contains(parts, "")
}

func generateTokenClaims(subject, id, scope, sessionID string, ttl time.Duration) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: coreutils.ParseJWTTimestamp(ttl),
		},
		Scope:     scope,
		SessionID: sessionID,
	})
}

// GenerateToken generates a JWT token with the specified subject, id, scope, and ttl
func (tm *tokenManager) GenerateToken(subject, id, scope string, ttl time.Duration) (*TokenResponse, *apperror.AppError) {
	token := generateTokenClaims(subject, id, scope, "", ttl)

	tokenStr, err := token.SignedString(tm.config.JwtSecret)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to sign token").
			WithDetail("redis_error", err.Error())
	}

	return &TokenResponse{TokenString: tokenStr, TokenID: id, Subject: subject}, nil
}

// MapTokenError maps a JWT error to an AppError
//
//   - Invalid Signature: CodeTokenInvalid
//   - Expired: CodeTokenExpired
//   - Malformed: CodeTokenMalformed
//   - Other: CodeTokenInvalid
func MapTokenError(err error) *apperror.AppError {
	switch {
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return apperror.New(apperror.CodeTokenInvalidSignature, "invalid token")
	case errors.Is(err, jwt.ErrTokenExpired):
		return apperror.New(apperror.CodeTokenExpired, "token has expired")
	case errors.Is(err, jwt.ErrTokenMalformed):
		return apperror.New(apperror.CodeTokenMalformed, "malformed token")
	default:
		// Don't expose internal error details to client
		return apperror.New(apperror.CodeTokenInvalid, "invalid token")
	}
}

func mapRedisError(err error) *apperror.AppError {
	if errors.Is(err, redis.Nil) {
		return apperror.New(apperror.CodeNotFound, "token not found")
	}

	// Distinguish between different Redis errors
	if errors.Is(err, context.DeadlineExceeded) {
		return apperror.New(apperror.CodeTimeout, "redis operation timed out")
	}

	// Log the actual error for debugging
	return apperror.ErrInternal.WithDetail("redis_error", err.Error())
}
