package token_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
)

func Test_RevokeAuthToken(t *testing.T) {
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: time.Minute,
		SessionTTL:     time.Minute,
	}, redisClient)

	ctx := context.Background()
	userID := "user123"

	tests := []struct {
		name      string
		setup     func(ctx context.Context, t *testing.T) (tokenID string, userID string)
		tokenID   string
		userID    string
		expectErr bool
	}{
		{
			name: "should successfully revoke existing token",
			setup: func(ctx context.Context, t *testing.T) (string, string) {
				tokenRes, appErr := tm.IssueAuthToken(ctx, userID, "")
				if appErr != nil {
					t.Fatalf("failed to issue token: %v", appErr.Message)
				}

				// Parse to get token ID
				jwtToken, err := jwt.ParseWithClaims(tokenRes.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
					return []byte("secret"), nil
				})
				if err != nil {
					t.Fatalf("failed to parse token: %v", err)
				}
				claims := jwtToken.Claims.(*token.Claims)
				return claims.ID, userID
			},
			expectErr: false,
		},
		{
			name: "should handle revoking non-existent token",
			setup: func(ctx context.Context, t *testing.T) (string, string) {
				return "non-existent-token-id", userID
			},
			expectErr: false, // Deleting non-existent key shouldn't error
		},
		{
			name: "should remove token from user's set",
			setup: func(ctx context.Context, t *testing.T) (string, string) {
				tokenStr, appErr := tm.IssueAuthToken(ctx, userID, "")
				if appErr != nil {
					t.Fatalf("failed to issue token: %v", appErr.Message)
				}

				jwtToken, err := jwt.ParseWithClaims(tokenStr.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
					return []byte("secret"), nil
				})
				if err != nil {
					t.Fatalf("failed to parse token: %v", err)
				}
				claims := jwtToken.Claims.(*token.Claims)
				return claims.ID, userID
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenID, userID := tt.setup(ctx, t)

			// Revoke the token
			appErr := tm.RevokeAuthToken(ctx, tokenID, userID)
			if tt.expectErr && appErr == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && appErr != nil {
				t.Fatalf("unexpected error: %v", appErr.Message)
			}

			// Verify token no longer exists in Redis
			val, err := redisClient.Get(ctx, tokenID).Result()
			if err != redis.Nil {
				t.Errorf("token should be deleted from Redis, but found value: %v", val)
			}

			// Verify token is removed from user's set
			userTokenKey := fmt.Sprintf("user_tokens:%s", userID)
			isMember, err := redisClient.SIsMember(ctx, userTokenKey, tokenID).Result()
			if err != nil {
				t.Fatalf("failed to check set membership: %v", err)
			}
			if isMember {
				t.Error("token ID should not be in user's set after revocation")
			}
		})
	}
}

func Test_RevokeAllUserTokens(t *testing.T) {
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: time.Minute,
		SessionTTL:     time.Minute,
	}, redisClient)

	ctx := context.Background()

	t.Run("should revoke all tokens for a user", func(t *testing.T) {
		userID := "user123"
		numTokens := 5
		tokenIDs := make([]string, numTokens)

		// Issue multiple tokens for the same user
		for i := range numTokens {
			tokenRes, appErr := tm.IssueAuthToken(ctx, userID, "")
			if appErr != nil {
				t.Fatalf("failed to issue token %d: %v", i, appErr.Message)
			}

			jwtToken, err := jwt.ParseWithClaims(tokenRes.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
				return []byte("secret"), nil
			})
			if err != nil {
				t.Fatalf("failed to parse token %d: %v", i, err)
			}
			claims := jwtToken.Claims.(*token.Claims)
			tokenIDs[i] = claims.ID
		}

		// Verify all tokens exist in Redis
		for i, tokenID := range tokenIDs {
			val, err := redisClient.Get(ctx, tokenID).Result()
			if err != nil {
				t.Fatalf("token %d should exist before revocation: %v", i, err)
			}
			if val != "active" {
				t.Errorf("token %d should be 'active', got '%s'", i, val)
			}
		}

		// Verify user's set exists
		userTokenKey := fmt.Sprintf("user_tokens:%s", userID)
		setSize, err := redisClient.SCard(ctx, userTokenKey).Result()
		if err != nil {
			t.Fatalf("failed to get set size: %v", err)
		}
		if setSize != int64(numTokens) {
			t.Errorf("expected set size %d, got %d", numTokens, setSize)
		}

		// Revoke all tokens
		appErr := tm.RevokeAllUserTokens(ctx, userID)
		if appErr != nil {
			t.Fatalf("unexpected error: %v", appErr.Message)
		}

		// Verify all tokens are deleted
		for i, tokenID := range tokenIDs {
			_, err := redisClient.Get(ctx, tokenID).Result()
			if err != redis.Nil {
				t.Errorf("token %d (%s) should be deleted, but still exists", i, tokenID)
			}
		}

		// Verify user's set is deleted
		exists, err := redisClient.Exists(ctx, userTokenKey).Result()
		if err != nil {
			t.Fatalf("failed to check set existence: %v", err)
		}
		if exists > 0 {
			t.Error("user token set should be deleted")
		}
	})

	t.Run("should handle user with no tokens", func(t *testing.T) {
		userID := "user_with_no_tokens"

		appErr := tm.RevokeAllUserTokens(ctx, userID)
		if appErr != nil {
			t.Fatalf("unexpected error for user with no tokens: %v", appErr.Message)
		}
	})

	t.Run("should not affect other users' tokens", func(t *testing.T) {
		user1 := "user1"
		user2 := "user2"

		// Issue tokens for both users
		token1Res, appErr := tm.IssueAuthToken(ctx, user1, "")
		if appErr != nil {
			t.Fatalf("failed to issue token for user1: %v", appErr.Message)
		}

		token2Res, appErr := tm.IssueAuthToken(ctx, user2, "")
		if appErr != nil {
			t.Fatalf("failed to issue token for user2: %v", appErr.Message)
		}

		// Parse to get token IDs
		jwtToken1, _ := jwt.ParseWithClaims(token1Res.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
			return []byte("secret"), nil
		})
		jwtToken2, _ := jwt.ParseWithClaims(token2Res.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
			return []byte("secret"), nil
		})
		claims1 := jwtToken1.Claims.(*token.Claims)
		claims2 := jwtToken2.Claims.(*token.Claims)

		// Revoke all tokens for user1 only
		appErr = tm.RevokeAllUserTokens(ctx, user1)
		if appErr != nil {
			t.Fatalf("unexpected error: %v", appErr.Message)
		}

		// Verify user1's token is deleted
		_, err := redisClient.Get(ctx, claims1.ID).Result()
		if err != redis.Nil {
			t.Error("user1's token should be deleted")
		}

		// Verify user2's token still exists
		val, err := redisClient.Get(ctx, claims2.ID).Result()
		if err != nil {
			t.Fatalf("user2's token should still exist: %v", err)
		}
		if val != "active" {
			t.Errorf("user2's token should be 'active', got '%s'", val)
		}

		// Verify user2's set still exists
		user2SetKey := fmt.Sprintf("user_tokens:%s", user2)
		isMember, err := redisClient.SIsMember(ctx, user2SetKey, claims2.ID).Result()
		if err != nil {
			t.Fatalf("failed to check set membership: %v", err)
		}
		if !isMember {
			t.Error("user2's token should still be in their set")
		}
	})
}

func Test_RevokeAuthToken_AfterRevokeAll(t *testing.T) {
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: time.Minute,
		SessionTTL:     time.Minute,
	}, redisClient)

	ctx := context.Background()
	userID := "user123"

	// Issue tokens
	tokenStrs := make([]string, 3)
	tokenIDs := make([]string, 3)

	for i := range 3 {
		tokenRes, appErr := tm.IssueAuthToken(ctx, userID, "")
		if appErr != nil {
			t.Fatalf("failed to issue token %d: %v", i, appErr.Message)
		}
		tokenStrs[i] = tokenRes.TokenString

		jwtToken, _ := jwt.ParseWithClaims(tokenRes.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
			return []byte("secret"), nil
		})
		claims := jwtToken.Claims.(*token.Claims)
		tokenIDs[i] = claims.ID
	}

	// Revoke all tokens
	appErr := tm.RevokeAllUserTokens(ctx, userID)
	if appErr != nil {
		t.Fatalf("failed to revoke all tokens: %v", appErr.Message)
	}

	// Try to revoke individual token after bulk revocation
	for _, tokenID := range tokenIDs {
		appErr := tm.RevokeAuthToken(ctx, tokenID, userID)
		if appErr != nil {
			t.Errorf("revoking already deleted token should not error: %v", appErr.Message)
		}
	}
}

func Test_RevokeAuthToken_Concurrent(t *testing.T) {
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: time.Minute,
		SessionTTL:     time.Minute,
	}, redisClient)

	ctx := context.Background()
	userID := "user123"

	// Issue a token
	tokenRes, appErr := tm.IssueAuthToken(ctx, userID, "")
	if appErr != nil {
		t.Fatalf("failed to issue token: %v", appErr.Message)
	}

	jwtToken, _ := jwt.ParseWithClaims(tokenRes.TokenString, &token.Claims{}, func(t *jwt.Token) (any, error) {
		return []byte("secret"), nil
	})
	claims := jwtToken.Claims.(*token.Claims)

	// Try to revoke the same token concurrently
	done := make(chan bool)
	errors := make(chan error, 5)

	for range 5 {
		go func() {
			err := tm.RevokeAuthToken(ctx, claims.ID, userID)
			if err != nil {
				errors <- fmt.Errorf("revoke error: %v", err.Message)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 5 {
		<-done
	}
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent revocation error: %v", err)
	}

	// Verify token is deleted (only once)
	_, err := redisClient.Get(ctx, claims.ID).Result()
	if err != redis.Nil {
		t.Error("token should be deleted after concurrent revocations")
	}
}

func Test_RevokeAllUserTokens_PartialFailure(t *testing.T) {
	// This test would require mocking Redis to simulate partial failures
	// For now, just test the success path
	t.Skip("Requires Redis mock for partial failure testing")
}
