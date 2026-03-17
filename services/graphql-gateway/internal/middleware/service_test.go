package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewareDoesNotRejectAnonymousRequests(t *testing.T) {
	m := &middleware{}
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
	m := &middleware{}

	handler := m.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Is-Authenticated"); got != "true" {
			t.Fatalf("expected authenticated header to be true, got %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}
