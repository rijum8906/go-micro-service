// Package jwt
package jwt

import (
	"context"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Issuer     string
	Secret     string
	Expiration time.Duration
}

type service struct {
	redis *redis.Client
	cfg   Config
}

func NewService(r *redis.Client, cfg Config) Service {
	return &service{
		redis: r,
		cfg:   cfg,
	}
}

func (s *service) IssueToken(ctx context.Context, subject string) (string, error) {
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
		return "", err
	}

	if err := s.redis.Set(
		ctx,
		"auth:session:"+sessionID,
		subject,
		time.Until(exp),
	).Err(); err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (s *service) ValidateToken(ctx context.Context, tokenStr string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(
		tokenStr,
		&jwtlib.RegisteredClaims{},
		func(t *jwtlib.Token) (any, error) {
			return []byte(s.cfg.Secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	rc := token.Claims.(*jwtlib.RegisteredClaims)

	exists, err := s.redis.Exists(ctx, "auth:session:"+rc.ID).Result()
	if err != nil || exists == 0 {
		return nil, err
	}

	return &Claims{
		UserID:    rc.Subject,
		SessionID: rc.ID,
		ExpiresAt: rc.ExpiresAt.Time,
	}, nil
}

func (s *service) RevokeSession(ctx context.Context, sessionID string) error {
	return s.redis.Del(ctx, "auth:session:"+sessionID).Err()
}
