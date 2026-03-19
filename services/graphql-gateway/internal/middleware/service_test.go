package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/rijum8906/relay/packages/common/env"
)

func TestAuthMiddlewareDoesNotRejectAnonymousRequests(t *testing.T) {
	m := &middleware{env: &env.Env{JwtSecret: "test-secret"}}
	called := false

	handler := m.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := r.Header.Get("X-Is-Authenticated"); got != "false" {
			t.Fatalf("expected anonymous header to be false, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestAuthMiddlewareMarksAuthenticatedRequests(t *testing.T) {
	m := &middleware{env: &env.Env{JwtSecret: "test-secret"}}

	handler := m.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Is-Authenticated"); got != "true" {
			t.Fatalf("expected authenticated header to be true, got %q", got)
		}
		if got := r.Header.Get("X-User-ID"); got != "user-123" {
			t.Fatalf("expected X-User-ID to be %q, got %q", "user-123", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Authorization", "Bearer "+signedTestToken(t, "test-secret", "user-123"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func signedTestToken(t *testing.T, secret, subject string) string {
	t.Helper()

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtlib.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	return signed
}
