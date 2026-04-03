package token

import (
	"context"
	"fmt"

	"github.com/dgryski/trifles/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rijum8906/relay/packages/core/apperror"
)

// NOTE: Auth Token Redis key format {UserID}:{SessionID}
// Scoped Token Redis key format {Token}

func (m *TokenManager) IssueAuthToken(ctx context.Context, subject, sessionID string, scope TokenScope) (string, *apperror.AppError) {
	// NOTE: use this same key for jti and redis key
	key := generateAuthTokenKey(subject, sessionID)

	tokenClaims := generateTokenClaims(subject, sessionID, scope, m.Config.SessionTTL)

	token, err := tokenClaims.SignedString(m.Config.JwtSecret)
	if err != nil {
		return "", apperror.New(apperror.CodeInternal, "failed to generate token").WithDetail("error", err.Error())
	}

	err = m.Store.Set(ctx, key, subject, m.Config.SessionTTL)
	if err != nil {
		return "", apperror.New(apperror.CodeInternal, "failed to set token").WithDetail("error", err.Error())
	}

	return token, nil
}

func (m *TokenManager) IssueScopedToken(ctx context.Context, subject string, scope TokenScope) (string, *apperror.AppError) {
	// NOTE: for scoped token, the jti will be random and the redis key will be the token
	tokenClaims := generateTokenClaims(subject, uuid.UUIDv4(), scope, m.Config.ScopedTokenTTL)

	token, err := tokenClaims.SignedString(m.Config.JwtSecret)
	if err != nil {
		return "", apperror.New(apperror.CodeInternal, "failed to generate token").WithDetail("error", err.Error())
	}

	err = m.Store.Set(ctx, token, subject, m.Config.ScopedTokenTTL)
	if err != nil {
		return "", apperror.New(apperror.CodeInternal, "failed to set token").WithDetail("error", err.Error())
	}

	return token, nil
}

func (m *TokenManager) ValidateAuthToken(ctx context.Context, tokenStr string) (*Claims, *apperror.AppError) {
	// 1. Parse & Verify JWT Signature/Expiration
	// This prevents hitting Redis for junk/malformed tokens.
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return m.Config.JwtSecret, nil
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

	key := generateAuthTokenKey(claims.Subject, sessionKey)

	// 3. Check Redis
	// If the key is gone, the session was revoked or naturally expired in Redis
	_, err = m.Store.Get(ctx, key)
	if err != nil {
		return nil, mapRedisError(err) // Uses your new mapper to return 401 if missing
	}

	return claims, nil
}

func (m *TokenManager) ValidateScopedToken(ctx context.Context, tokenStr string) (*Claims, *apperror.AppError) {
	// 1. Parse & Verify JWT Signature/Expiration first
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return m.Config.JwtSecret, nil
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
	_, err = m.Store.Get(ctx, tokenStr)
	if err != nil {
		return nil, mapRedisError(err) // Uses your new mapper to return 401 if missing
	}

	return claims, nil
}

func (m *TokenManager) RevokeAuthToken(ctx context.Context, subject, sessionID string) *apperror.AppError {
	key := generateAuthTokenKey(subject, sessionID)

	err := m.Store.Del(ctx, key)
	if err != nil {
		return apperror.New(apperror.CodeInternal, "failed to revoke token").WithDetail("error", err.Error())
	}
	return nil
}

func (m *TokenManager) RevokeScopedToken(ctx context.Context, token string) *apperror.AppError {
	if err := m.Store.Del(ctx, token); err != nil {
		return apperror.New(apperror.CodeInternal, "failed to revoke token").WithDetail("error", err.Error())
	}

	return nil
}

func (m *TokenManager) RevokeAllUserTokens(ctx context.Context, subject string) *apperror.AppError {
	// Find all keys starting with "userID:*"
	pattern := fmt.Sprintf("%s:*", subject)

	// Scan and delete all matching session keys
	var cursor uint64
	for {
		keys, nextCursor, err := m.Store.Scan(ctx, cursor, pattern, 10)
		if err != nil {
			return apperror.ErrInternal.WithMessage("failed to find sessions")
		}

		if len(keys) > 0 {
			m.Store.Del(ctx, keys...)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (m *TokenManager) RevokeOtherUserTokens(ctx context.Context, subject, currentSessionID string) *apperror.AppError {
	pattern := fmt.Sprintf("%s:*", subject)
	currentKey := generateAuthTokenKey(subject, currentSessionID)

	var cursor uint64
	for {
		keys, nextCursor, err := m.Store.Scan(ctx, cursor, pattern, 10)
		if err != nil {
			return apperror.ErrInternal.WithMessage("failed to find sessions")
		}

		filtered := make([]string, 0, len(keys))
		for _, key := range keys {
			if key == currentKey {
				continue
			}
			filtered = append(filtered, key)
		}

		if len(filtered) > 0 {
			m.Store.Del(ctx, filtered...)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
