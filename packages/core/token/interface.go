// Package token contains test utils for token
package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	JwtSecret      []byte
	SessionTTL     time.Duration
	ScopedSecret   []byte
	ScopedTokenTTL time.Duration
}

// Claims is a custom claims struct that includes the scope of the token
type Claims struct {
	Scope     string `json:"scope"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// TokenManager is a struct that holds the configuration and methods for issuing, validating and revoking tokens
type TokenManager struct {
	// Core
	redisClient *redis.Client

	// config
	config Config
}

// NewTokenManager returns a new instance of TokenManager
func NewTokenManager(config Config, redisClient *redis.Client) *TokenManager {
	return &TokenManager{
		config:      config,
		redisClient: redisClient,
	}
}

// SignedString returns a signed string of the token
func (c *Claims) SignedString(secret []byte) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(secret)
}
