package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
)

func Test_IssueScopedToken(t *testing.T) {
	// Setup
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: time.Minute,
		SessionTTL:     time.Minute,
	}, redisClient)

	ctx := context.Background()
	subject := "user@example.com"

	tests := []struct {
		name        string
		scope       string
		expectError bool
	}{
		{
			name:        "should issue email verification token",
			scope:       "email_verification",
			expectError: false,
		},
		{
			name:        "should issue password reset token",
			scope:       "password_reset",
			expectError: false,
		},
		{
			name:        "should issue invite token",
			scope:       "invite",
			expectError: false,
		},
		{
			name:        "should handle magic link scope",
			scope:       "magic_link",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Issue token
			tokenRes, appErr := tm.IssueScopedToken(ctx, subject, tt.scope)
			if tt.expectError {
				if appErr == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if appErr != nil {
				t.Fatalf("unexpected error: %v", appErr.Message)
			}

			// Parse and validate JWT
			jwtToken, err := jwt.ParseWithClaims(tokenRes.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
				return []byte("scoped-secret"), nil
			})
			if err != nil {
				t.Fatalf("failed to parse JWT: %v", err)
			}

			claims, ok := jwtToken.Claims.(*token.Claims)
			if !ok {
				t.Fatal("failed to cast token claims")
			}

			// Verify claims
			if claims.Scope != tt.scope {
				t.Errorf("expected scope %s, got %s", tt.scope, claims.Scope)
			}

			if claims.Subject != subject {
				t.Errorf("expected subject %s, got %s", subject, claims.Subject)
			}

			// Verify token ID is a valid UUID
			if _, err := uuid.Parse(claims.ID); err != nil {
				t.Errorf("token ID is not a valid UUID: %s", claims.ID)
			}

			// Verify expiration is set
			if claims.ExpiresAt == nil {
				t.Error("expected expiration to be set")
			}

			// Verify Redis storage
			val, err := redisClient.Get(ctx, claims.ID).Result()
			if err != nil {
				t.Fatalf("failed to get token from Redis: %v", err)
			}
			if val != "active" {
				t.Errorf("expected Redis value 'active', got '%s'", val)
			}

			// Verify Redis TTL
			ttl, err := redisClient.TTL(ctx, claims.ID).Result()
			if err != nil {
				t.Fatalf("failed to get TTL: %v", err)
			}
			if ttl <= 0 || ttl > time.Minute {
				t.Errorf("expected TTL between 0 and 1 minute, got %v", ttl)
			}
		})
	}
}

func Test_IssueScopedToken_Uniqueness(t *testing.T) {
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: time.Minute,
		SessionTTL:     time.Minute,
	}, redisClient)

	ctx := context.Background()
	subject := "user@example.com"
	scope := "email_verification"

	// Issue multiple tokens
	tokens := make([]string, 5)
	tokenIDs := make([]string, 5)

	for i := range 5 {
		tokenRes, appErr := tm.IssueScopedToken(ctx, subject, scope)
		if appErr != nil {
			t.Fatalf("failed to issue token %d: %v", i, appErr.Message)
		}
		tokens[i] = tokenRes.TokenString

		// Parse to get token ID
		jwtToken, err := jwt.ParseWithClaims(tokenRes.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
			return []byte("scoped-secret"), nil
		})
		if err != nil {
			t.Fatalf("failed to parse token %d: %v", i, err)
		}
		claims := jwtToken.Claims.(*token.Claims)
		tokenIDs[i] = claims.ID
	}

	// Verify all tokens are unique
	uniqueTokens := make(map[string]bool)
	uniqueTokenIDs := make(map[string]bool)

	for i, tokenStr := range tokens {
		if uniqueTokens[tokenStr] {
			t.Errorf("duplicate token string found at index %d", i)
		}
		uniqueTokens[tokenStr] = true

		if uniqueTokenIDs[tokenIDs[i]] {
			t.Errorf("duplicate token ID found at index %d: %s", i, tokenIDs[i])
		}
		uniqueTokenIDs[tokenIDs[i]] = true
	}

	if len(uniqueTokens) != 5 {
		t.Errorf("expected 5 unique tokens, got %d", len(uniqueTokens))
	}
}

func Test_IssueScopedToken_RedisExpiration(t *testing.T) {
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	shortTTL := 2 * time.Second
	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: shortTTL,
		SessionTTL:     time.Minute,
	}, redisClient)

	ctx := context.Background()

	// Issue token
	tokenRes, appErr := tm.IssueScopedToken(ctx, "user@example.com", "email_verification")
	if appErr != nil {
		t.Fatalf("failed to issue token: %v", appErr.Message)
	}

	// Parse to get claims
	jwtToken, err := jwt.ParseWithClaims(tokenRes.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
		return []byte("scoped-secret"), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	claims := jwtToken.Claims.(*token.Claims)

	// Verify token exists in Redis
	_, err = redisClient.Get(ctx, claims.ID).Result()
	if err != nil {
		t.Fatalf("token should exist in Redis: %v", err)
	}

	// Wait for expiration
	time.Sleep(shortTTL + 1*time.Second)

	// Verify token is automatically removed from Redis
	_, err = redisClient.Get(ctx, claims.ID).Result()
	if err != redis.Nil {
		t.Errorf("expected token to be expired from Redis, got error: %v", err)
	}
}

func Test_IssueScopedToken_ContextCancellation(t *testing.T) {
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: time.Minute,
		SessionTTL:     time.Minute,
	}, redisClient)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tokenRes, appErr := tm.IssueScopedToken(ctx, "user@example.com", "email_verification")
	if appErr == nil {
		t.Fatal("expected error with cancelled context, got none")
	}
	if tokenRes != nil {
		t.Errorf("expected empty token string, got %s", tokenRes)
	}
}
