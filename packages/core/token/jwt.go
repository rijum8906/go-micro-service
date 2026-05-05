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
	Role  string     `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenManager(config Config, redisClient *redis.Client) *TokenManager {
	store := NewRedisStore(redisClient)
	return &TokenManager{
		Config: config,
		Store:  store,
	}
}
