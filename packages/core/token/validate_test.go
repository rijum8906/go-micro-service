package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
)

func Test_ValidateAuthToken(t *testing.T) {
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

	t.Run("should validate valid auth token", func(t *testing.T) {
		// Issue a token
		tokenRes, appErr := tm.IssueAuthToken(ctx, userID, "")
		if appErr != nil {
			t.Fatalf("failed to issue token: %v", appErr.Message)
		}

		// Validate the token
		claims, appErr := tm.ValidateAuthToken(ctx, tokenRes.TokenString)
		if appErr != nil {
			t.Fatalf("unexpected error: %v", appErr.Message)
		}

		// Verify claims
		if claims.Subject != userID {
			t.Errorf("expected subject %s, got %s", userID, claims.Subject)
		}
		if claims.Scope != token.TokenScopeAuth {
			t.Errorf("expected scope %s, got %s", token.TokenScopeAuth, claims.Scope)
		}
	})

	t.Run("should reject token with wrong secret", func(t *testing.T) {
		// Create a token with wrong secret (simulate by creating manually)
		wrongClaims := &token.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID,
				ID:        uuid.NewString(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			},
			Scope: token.TokenScopeAuth,
		}

		tokenStr, err := wrongClaims.SignedString([]byte("wrong-secret"))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, appErr := tm.ValidateAuthToken(ctx, tokenStr)
		if appErr == nil {
			t.Fatal("expected error for wrong secret, got none")
		}
		if appErr.Code != apperror.CodeTokenInvalidSignature {
			t.Errorf("expected CodeTokenInvalidSignature, got %s", appErr.Code)
		}
	})

	t.Run("should reject expired token", func(t *testing.T) {
		// Create expired token
		expiredClaims := &token.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID,
				ID:        uuid.NewString(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			},
			Scope: token.TokenScopeAuth,
		}

		tokenStr, err := expiredClaims.SignedString([]byte("secret"))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, appErr := tm.ValidateAuthToken(ctx, tokenStr)
		if appErr == nil {
			t.Fatal("expected error for expired token, got none")
		}
		if appErr.Code != apperror.CodeTokenExpired {
			t.Errorf("expected CodeTokenExpired, got %s", appErr.Code)
		}
	})

	t.Run("should reject revoked token", func(t *testing.T) {
		// Issue a token
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

		// Revoke the token
		appErr = tm.RevokeAuthToken(ctx, claims.ID, userID)
		if appErr != nil {
			t.Fatalf("failed to revoke token: %v", appErr.Message)
		}

		// Try to validate revoked token
		_, appErr = tm.ValidateAuthToken(ctx, tokenRes.TokenString)
		if appErr == nil {
			t.Fatal("expected error for revoked token, got none")
		}
		if appErr.Code != apperror.CodeTokenExpired {
			t.Errorf("expected CodeTokenExpired, got %s", appErr.Code)
		}
	})

	t.Run("should reject scoped token as auth token", func(t *testing.T) {
		// Issue a scoped token
		scopedTokenRes, appErr := tm.IssueScopedToken(ctx, userID, "email_verification")
		if appErr != nil {
			t.Fatalf("failed to issue scoped token: %v", appErr.Message)
		}

		// Try to validate as auth token
		_, appErr = tm.ValidateAuthToken(ctx, scopedTokenRes.TokenString)
		if appErr == nil {
			t.Fatal("expected error for scoped token used as auth token, got none")
		}
		if appErr.Code != apperror.CodeTokenInvalidSignature {
			t.Errorf("expected CodeTokenInvalid, got %s", appErr.Code)
		}
	})

	t.Run("should reject malformed token", func(t *testing.T) {
		_, appErr := tm.ValidateAuthToken(ctx, "invalid-token-string")
		if appErr == nil {
			t.Fatal("expected error for malformed token, got none")
		}
		if appErr.Code != apperror.CodeTokenMalformed {
			t.Errorf("expected CodeTokenMalformed, got %s", appErr.Code)
		}
	})
}

func Test_ValidateScopedToken(t *testing.T) {
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

	t.Run("should validate valid scoped token once", func(t *testing.T) {
		// Issue a scoped token
		tokenRes, appErr := tm.IssueScopedToken(ctx, subject, scope)
		if appErr != nil {
			t.Fatalf("failed to issue token: %v", appErr.Message)
		}

		// First validation should succeed
		claims, appErr := tm.ValidateScopedToken(ctx, tokenRes.TokenString)
		if appErr != nil {
			t.Fatalf("unexpected error on first validation: %v", appErr.Message)
		}

		// Verify claims
		if claims.Subject != subject {
			t.Errorf("expected subject %s, got %s", subject, claims.Subject)
		}
		if claims.Scope != scope {
			t.Errorf("expected scope %s, got %s", scope, claims.Scope)
		}

		// Second validation should fail (token already used)
		_, appErr = tm.ValidateScopedToken(ctx, tokenRes.TokenString)
		if appErr == nil {
			t.Fatal("expected error on second validation, got none")
		}
		if appErr.Code != apperror.CodeTokenExpired {
			t.Errorf("expected CodeTokenExpired, got %s", appErr.Code)
		}
	})

	t.Run("should reject token with wrong secret", func(t *testing.T) {
		// Create token with wrong secret
		wrongClaims := &token.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   subject,
				ID:        uuid.NewString(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			},
			Scope: scope,
		}

		tokenStr, err := wrongClaims.SignedString([]byte("wrong-secret"))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, appErr := tm.ValidateScopedToken(ctx, tokenStr)
		if appErr == nil {
			t.Fatal("expected error for wrong secret, got none")
		}
		if appErr.Code != apperror.CodeTokenInvalidSignature {
			t.Errorf("expected CodeTokenInvalidSignature, got %s", appErr.Code)
		}
	})

	t.Run("should reject expired scoped token", func(t *testing.T) {
		// Create expired token with correct secret
		expiredClaims := &token.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   subject,
				ID:        uuid.NewString(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			},
			Scope: scope,
		}

		tokenStr, err := expiredClaims.SignedString([]byte("scoped-secret"))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		// Store in Redis (simulate issuance)
		redisClient.Set(ctx, expiredClaims.ID, "active", time.Minute)

		_, appErr := tm.ValidateScopedToken(ctx, tokenStr)
		if appErr == nil {
			t.Fatal("expected error for expired token, got none")
		}
		if appErr.Code != apperror.CodeTokenExpired {
			t.Errorf("expected CodeTokenExpired, got %s", appErr.Code)
		}
	})

	t.Run("should reject already used token", func(t *testing.T) {
		// Issue a token
		tokenRes, appErr := tm.IssueScopedToken(ctx, subject, scope)
		if appErr != nil {
			t.Fatalf("failed to issue token: %v", appErr.Message)
		}

		// Use it once
		_, appErr = tm.ValidateScopedToken(ctx, tokenRes.TokenString)
		if appErr != nil {
			t.Fatalf("first validation failed: %v", appErr.Message)
		}

		// Try to use again
		_, appErr = tm.ValidateScopedToken(ctx, tokenRes.TokenString)
		if appErr == nil {
			t.Fatal("expected error on second use, got none")
		}
		if appErr.Code != apperror.CodeTokenExpired {
			t.Errorf("expected CodeTokenExpired, got %s", appErr.Code)
		}
	})

	t.Run("should reject auth token as scoped token", func(t *testing.T) {
		// Issue an auth token
		authTokenRes, appErr := tm.IssueAuthToken(ctx, subject, "")
		if appErr != nil {
			t.Fatalf("failed to issue auth token: %v", appErr.Message)
		}

		// Try to validate as scoped token
		_, appErr = tm.ValidateScopedToken(ctx, authTokenRes.TokenString)
		if appErr == nil {
			t.Fatal("expected error for auth token used as scoped token, got none")
		}
		// Should fail because secret mismatch (auth token signed with JwtSecret, not ScopedSecret)
		if appErr.Code != apperror.CodeTokenInvalidSignature {
			t.Errorf("expected CodeTokenInvalidSignature, got %s", appErr.Code)
		}
	})

	t.Run("should handle concurrent validation of same token", func(t *testing.T) {
		// Issue a token
		tokenRes, appErr := tm.IssueScopedToken(ctx, subject, scope)
		if appErr != nil {
			t.Fatalf("failed to issue token: %v", appErr.Message)
		}

		// Try to validate concurrently
		const goroutines = 10
		results := make(chan bool, goroutines)

		for range goroutines {
			go func() {
				_, appErr := tm.ValidateScopedToken(ctx, tokenRes.TokenString)
				results <- appErr == nil
			}()
		}

		// Count successes
		successCount := 0
		for range goroutines {
			if <-results {
				successCount++
			}
		}

		// Only one should succeed (atomic GetDel)
		if successCount != 1 {
			t.Errorf("expected exactly 1 successful validation, got %d", successCount)
		}
	})

	t.Run("should reject malformed scoped token", func(t *testing.T) {
		_, appErr := tm.ValidateScopedToken(ctx, "invalid-token")
		if appErr == nil {
			t.Fatal("expected error for malformed token, got none")
		}
		if appErr.Code != apperror.CodeTokenMalformed {
			t.Errorf("expected CodeTokenMalformed, got %s", appErr.Code)
		}
	})

	t.Run("should reject token with invalid state in Redis", func(t *testing.T) {
		// Create valid JWT
		claims := &token.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   subject,
				ID:        uuid.NewString(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			},
			Scope: scope,
		}

		tokenStr, err := claims.SignedString([]byte("scoped-secret"))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		// Store invalid state in Redis
		redisClient.Set(ctx, claims.ID, "invalid-state", time.Minute)

		// Validate should fail
		_, appErr := tm.ValidateScopedToken(ctx, tokenStr)
		if appErr == nil {
			t.Fatal("expected error for invalid token state, got none")
		}
		if appErr.Code != apperror.CodeTokenInvalid {
			t.Errorf("expected CodeTokenInvalid, got %s", appErr.Code)
		}
	})

	t.Run("should reject token not found in Redis", func(t *testing.T) {
		// Create valid JWT but never store in Redis
		claims := &token.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   subject,
				ID:        uuid.NewString(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			},
			Scope: scope,
		}

		tokenStr, err := claims.SignedString([]byte("scoped-secret"))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		// Validate should fail (token not in Redis)
		_, appErr := tm.ValidateScopedToken(ctx, tokenStr)
		if appErr == nil {
			t.Fatal("expected error for token not in Redis, got none")
		}
		if appErr.Code != apperror.CodeTokenExpired {
			t.Errorf("expected CodeTokenExpired, got %s", appErr.Code)
		}
	})
}

func Test_ValidateAuthToken_WithContextCancellation(t *testing.T) {
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: time.Minute,
		SessionTTL:     time.Minute,
	}, redisClient)

	// Issue a token
	tokenRes, appErr := tm.IssueAuthToken(context.Background(), "user123", "")
	if appErr != nil {
		t.Fatalf("failed to issue token: %v", appErr.Message)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Validate with cancelled context
	_, appErr = tm.ValidateAuthToken(ctx, tokenRes.TokenString)
	if appErr == nil {
		t.Fatal("expected error with cancelled context, got none")
	}
}

func Test_ValidateScopedToken_WithContextCancellation(t *testing.T) {
	redisClient := testutils.MustConnectRedis()
	defer redisClient.FlushAll(context.Background())

	tm := token.NewTokenManager(token.Config{
		JwtSecret:      []byte("secret"),
		ScopedSecret:   []byte("scoped-secret"),
		ScopedTokenTTL: time.Minute,
		SessionTTL:     time.Minute,
	}, redisClient)

	// Issue a token
	tokenRes, appErr := tm.IssueScopedToken(context.Background(), "user@example.com", "email_verification")
	if appErr != nil {
		t.Fatalf("failed to issue token: %v", appErr.Message)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Validate with cancelled context
	_, appErr = tm.ValidateScopedToken(ctx, tokenRes.TokenString)
	if appErr == nil {
		t.Fatal("expected error with cancelled context, got none")
	}
}

func Test_ValidateAuthToken_RedisUnavailable(t *testing.T) {
	// This test would require mocking Redis
	// For now, skip as it needs a mock
	t.Skip("Requires Redis mock for testing unavailable scenario")
}

func Test_ValidateScopedToken_RedisUnavailable(t *testing.T) {
	// This test would require mocking Redis
	// For now, skip as it needs a mock
	t.Skip("Requires Redis mock for testing unavailable scenario")
}
