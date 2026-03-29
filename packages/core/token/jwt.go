// Package token provides JWT token logic
package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type redisStore struct {
	client *redis.Client
}

type Config struct {
	JwtSecret      []byte
	SessionTTL     time.Duration
	ScopedSecret   []byte
	ScopedTokenTTL time.Duration
}

type TokenManager struct {
	Config Config
	Store  TokenStore
}

type Claims struct {
	Scope TokenScope `json:"scope"`
	jwt.RegisteredClaims
}

const (
	defaultSessionTTL    = 15 * time.Minute
	defaultScopedTokenTTL = 10 * time.Minute
)

func NewTokenManager(jwtSec, scopedSec string, redisClient *redis.Client) *TokenManager {
	store := NewRedisStore(redisClient)
	return &TokenManager{
		Config: Config{
			JwtSecret:      []byte(jwtSec),
			ScopedSecret:   []byte(scopedSec),
			SessionTTL:     defaultSessionTTL,
			ScopedTokenTTL: defaultScopedTokenTTL,
		},
		Store: store,
	}
}
