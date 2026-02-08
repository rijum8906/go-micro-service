// Package jwt
package jwt

import (
	"context"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
)

func (s *service) IssueToken(
	ctx context.Context,
	subject string,
) (string, *errors.AppError) {
	sessionID := uuid.NewString()
	now := time.Now()
	exp := now.Add(s.cfg.Expiration)

	rc := jwtlib.RegisteredClaims{
		Subject:   subject,
		ID:        sessionID,
		IssuedAt:  jwtlib.NewNumericDate(now),
		ExpiresAt: jwtlib.NewNumericDate(exp),
		Issuer:    s.cfg.Issuer,
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, rc)

	tokenStr, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", errors.ErrInternal.WithInternal(err)
	}

	if err := s.redis.Set(
		ctx,
		sessionKey(sessionID),
		subject,
		time.Until(exp),
	).Err(); err != nil {
		return "", errors.ErrDBError.WithInternal(err)
	}

	return tokenStr, nil
}

func (s *service) ValidateToken(
	ctx context.Context,
	tokenStr string,
) (*Claims, *errors.AppError) {
	token, err := jwtlib.ParseWithClaims(
		tokenStr,
		&jwtlib.RegisteredClaims{},
		func(t *jwtlib.Token) (any, error) {
			return []byte(s.cfg.Secret), nil
		},
	)
	if err != nil || !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	rc, ok := token.Claims.(*jwtlib.RegisteredClaims)
	if !ok {
		return nil, errors.ErrInvalidTokenClaims
	}

	exists, err := s.redis.Exists(ctx, sessionKey(rc.ID)).Result()
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}
	if exists == 0 {
		return nil, errors.NewAppError(400, "session not found", []errors.Error{})
	}

	return &Claims{
		UserID:    rc.Subject,
		SessionID: rc.ID,
		ExpiresAt: rc.ExpiresAt.Time,
	}, nil
}

func (s *service) RevokeSession(
	ctx context.Context,
	sessionID string,
) *errors.AppError {
	err := s.redis.Del(ctx, sessionKey(sessionID)).Err()
	if err != nil {
		return errors.ErrDBError.WithInternal(err)
	}
	return nil
}
