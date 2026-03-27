// Package token provides JWT token logic
package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	jwtSecret      []byte
	SessionTTL     time.Duration
	scopedSecret   []byte
	ScopedTokenTTL time.Duration
}

type TokenManager struct {
	config Config
	redis  *redis.Client
}

type Claims struct {
	Scope TokenScope `json:"scope"`
	jwt.RegisteredClaims
}

func NewTokenManager(jwtSec, scopedSec string, redisClient *redis.Client) *TokenManager {
	return &TokenManager{
		config: Config{
			jwtSecret:    []byte(jwtSec),
			scopedSecret: []byte(scopedSec),
		},
		redis: redisClient,
	}
}
