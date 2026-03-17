package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/common/errors"
)

type Service interface {
	IssueToken(ctx context.Context, userID string) (string, *errors.AppError)
	ValidateToken(ctx context.Context, token string) (*Claims, *errors.AppError)
	RevokeSession(ctx context.Context, sessionID string) *errors.AppError
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

type Config struct {
	Issuer     string
	Secret     string
	Expiration time.Duration
}

type ScopedActionClaims struct {
	Subject string
	Scope   string
}

type ScopedJWTClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope"`
}

type ScopedActionJWT interface {
	IssueToken(ctx context.Context, claims ScopedActionClaims) (string, *errors.AppError)
	ValidateToken(ctx context.Context, token string) (*ScopedActionClaims, *errors.AppError)
	RevokeActionToken(ctx context.Context, jti string) *errors.AppError
}

type scopedActionJWT struct {
	redis *redis.Client
	cfg   Config
}

func NewScopedActionJWT(r *redis.Client, cfg Config) ScopedActionJWT {
	return &scopedActionJWT{
		redis: r,
		cfg:   cfg,
	}
}
