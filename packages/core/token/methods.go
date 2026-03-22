package token

import (
	"context"
	"fmt"

	"github.com/dgryski/trifles/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rijum8906/relay/packages/core/apperror"
)

func (m *TokenManager) IssueAuthToken(ctx context.Context, subject, sessionID, deviceID string, scope TokenScope) (string, *apperror.AppError) {
	// NOTE: use this same key for jti and redis key
	key := generateAuthTokenKey(subject, sessionID, deviceID)

	tokenClaims := generateTokenClaims(subject, key, scope, m.config.SessionTTL)

	token, err := tokenClaims.SignedString(m.config.jwtSecret)
	if err != nil {
		return "", apperror.New(apperror.TypeInternal, apperror.CodeInternal, "failed to generate token").WithDetail("error", err.Error())
	}

	_, err = m.redis.Set(ctx, key, subject, m.config.ScopedTokenTTL).Result()
	if err != nil {
		return "", apperror.New(apperror.TypeInternal, apperror.CodeInternal, "failed to set token").WithDetail("error", err.Error())
	}

	return token, nil
}

func (m *TokenManager) IssueScopedToken(ctx context.Context, subject string, scope TokenScope) (string, *apperror.AppError) {
	// NOTE: for scoped token, the jti will be random and the redis key will be the token
	tokenClaims := generateTokenClaims(subject, uuid.UUIDv4(), scope, m.config.ScopedTokenTTL)

	token, err := tokenClaims.SignedString(m.config.jwtSecret)
	if err != nil {
		return "", apperror.New(apperror.TypeInternal, apperror.CodeInternal, "failed to generate token").WithDetail("error", err.Error())
	}

	_, err = m.redis.Set(ctx, token, subject, m.config.ScopedTokenTTL).Result()
	if err != nil {
		return "", apperror.New(apperror.TypeInternal, apperror.CodeInternal, "failed to set token").WithDetail("error", err.Error())
	}

	return token, nil
}

func (m *TokenManager) ValidateAuthToken(ctx context.Context, tokenStr string) (*Claims, *apperror.AppError) {
	// 1. Parse & Verify JWT Signature/Expiration
	// This prevents hitting Redis for junk/malformed tokens.
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return m.config.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, apperror.ErrUnAuthenticated.WithMessage("invalid or expired token signature")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("failed to parse token claims")
	}

	// 2. Extract the composite key from the JWT 'ID' (jti)
	// NOTE: IssueAuthToken stored generateAuthTokenKey inside the ID
	sessionKey := claims.ID

	// 3. Check Redis
	// If the key is gone, the session was revoked or naturally expired in Redis
	_, err = m.redis.Get(ctx, sessionKey).Result()
	if err != nil {
		return nil, mapRedisError(err) // Uses your new mapper to return 401 if missing
	}

	return claims, nil
}

func (m *TokenManager) ValidateScopedToken(ctx context.Context, tokenStr string) (*Claims, *apperror.AppError) {
	// 1. Parse & Verify JWT Signature/Expiration first
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return m.config.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, apperror.ErrUnAuthenticated.WithMessage("invalid or expired token signature")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, apperror.ErrInternal.WithMessage("failed to parse token claims")
	}

	// 2. Check Redis
	// If the key is gone, the session was revoked or naturally expired in Redis
	_, err = m.redis.Get(ctx, tokenStr).Result()
	if err != nil {
		return nil, mapRedisError(err) // Uses your new mapper to return 401 if missing
	}

	return claims, nil
}

func (m *TokenManager) RevokeAuthToken(ctx context.Context, subject, sessionID, deviceID string) *apperror.AppError {
	key := generateAuthTokenKey(subject, sessionID, deviceID)

	_, err := m.redis.Del(ctx, key).Result()
	if err != nil {
		return apperror.New(apperror.TypeInternal, apperror.CodeInternal, "failed to revoke token").WithDetail("error", err.Error())
	}
	return nil
}

func (m *TokenManager) RevokeScopedToken(ctx context.Context, token string) *apperror.AppError {
	if err := m.redis.Del(ctx, token).Err(); err != nil {
		return apperror.New(apperror.TypeInternal, apperror.CodeInternal, "failed to revoke token").WithDetail("error", err.Error())
	}

	return nil
}

func (m *TokenManager) RevokeAllUserTokens(ctx context.Context, subject string) *apperror.AppError {
	// Find all keys starting with "userID:*"
	pattern := fmt.Sprintf("%s:*", subject)

	// Scan and delete all matching session keys
	var cursor uint64
	for {
		keys, nextCursor, err := m.redis.Scan(ctx, cursor, pattern, 10).Result()
		if err != nil {
			return apperror.ErrInternal.WithMessage("failed to find sessions")
		}

		if len(keys) > 0 {
			m.redis.Del(ctx, keys...)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
