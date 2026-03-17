package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/common/errors"
)

const (
	ActionChangePassword = "change_password"
	ActionChangeEmail    = "change_email"
	ActionChangeName     = "change_name"
	ActionChangePhone    = "change_phone"
	ActionChangeRole     = "change_role"
	ActionChangeStatus   = "change_status"
)

func (s *scopedActionJWT) IssueToken(
	ctx context.Context,
	claims ScopedActionClaims,
) (string, *errors.AppError) {
	jti := uuid.NewString()

	jwtClaims := ScopedJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   claims.Subject,
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.Expiration)),
		},
		Scope: claims.Scope,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenStr, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", errors.ErrInternal.WithInternal(err)
	}

	if err := s.redis.Set(
		ctx,
		redisKey(jti),
		claims.Scope,
		time.Until(jwtClaims.ExpiresAt.Time),
	).Err(); err != nil {
		return "", errors.ErrDBError.WithInternal(err)
	}

	return tokenStr, nil
}

func (s *scopedActionJWT) ValidateToken(
	ctx context.Context,
	tokenStr string,
) (*ScopedActionClaims, *errors.AppError) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&ScopedJWTClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(s.cfg.Secret), nil
		},
	)
	if err != nil || !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*ScopedJWTClaims)
	if !ok {
		return nil, errors.ErrInvalidTokenClaims
	}

	key := redisKey(claims.ID)

	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil || exists == 0 {
		return nil, errors.ErrInvalidToken
	}

	// single-use: burn after validation
	err = s.redis.Del(ctx, key).Err()
	if err != nil {
		return nil, errors.ErrDBError.WithInternal(err)
	}

	return &ScopedActionClaims{
		Subject: claims.Subject,
		Scope:   claims.Scope,
	}, nil
}

func (s *scopedActionJWT) RevokeActionToken(
	ctx context.Context,
	jti string,
) *errors.AppError {
	err := s.redis.Del(ctx, redisKey(jti)).Err()
	if err != nil {
		return errors.ErrDBError.WithInternal(err)
	}
	return nil
}
