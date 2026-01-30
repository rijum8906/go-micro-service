package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	ActionChangePassword = "change_password"
	ActionChangeEmail    = "change_email"
	ActionChangeName     = "change_name"
)

var ErrInvalidToken = errors.New("invalid or expired action token")

type ScopedActionClaims struct {
	UserID string
	Scope  string
	JTI    string
}

type ScopedJWTClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope"`
}

type ScopedActionJWT interface {
	IssueToken(ctx context.Context, claims ScopedActionClaims) (string, error)
	ValidateToken(ctx context.Context, token string) (*ScopedActionClaims, error)
	RevokeActionToken(ctx context.Context, jti string) error
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

func redisKey(jti string) string {
	return "scoped_action:" + jti
}

func (s *scopedActionJWT) IssueToken(
	ctx context.Context,
	claims ScopedActionClaims,
) (string, error) {
	jti := uuid.NewString()

	jwtClaims := ScopedJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   claims.UserID,
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.Expiration)),
		},
		Scope: claims.Scope,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenStr, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", err
	}

	if err := s.redis.Set(
		ctx,
		redisKey(jti),
		claims.Scope,
		time.Until(jwtClaims.ExpiresAt.Time),
	).Err(); err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (s *scopedActionJWT) ValidateToken(
	ctx context.Context,
	tokenStr string,
) (*ScopedActionClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&ScopedJWTClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(s.cfg.Secret), nil
		},
	)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*ScopedJWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	key := redisKey(claims.ID)

	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil || exists == 0 {
		return nil, ErrInvalidToken
	}

	// single-use: burn after validation
	_ = s.redis.Del(ctx, key)

	return &ScopedActionClaims{
		UserID: claims.Subject,
		Scope:  claims.Scope,
		JTI:    claims.ID,
	}, nil
}

func (s *scopedActionJWT) RevokeActionToken(
	ctx context.Context,
	jti string,
) error {
	return s.redis.Del(ctx, redisKey(jti)).Err()
}
