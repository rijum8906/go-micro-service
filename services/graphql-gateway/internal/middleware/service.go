package middleware

import (
	"net/http"
	"strings"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func (m *middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if userID, ok := m.userIDFromAuthorization(authorization); ok {
			r.Header.Set("X-Is-Authenticated", "true")
			r.Header.Set("X-User-ID", userID)
		} else {
			r.Header.Del("X-User-ID")
			r.Header.Set("X-Is-Authenticated", "false")
		}

		ipAddr := r.Header.Get("X-Forwarded-For")
		if ipAddr == "" {
			ipAddr = r.RemoteAddr
		}
		r.Header.Set("X-Client-IP", ipAddr)

		next.ServeHTTP(w, r)
	})
}

func (m *middleware) userIDFromAuthorization(authorization string) (string, bool) {
	if m == nil || m.env == nil || m.env.JwtSecret == "" {
		return "", false
	}

	tokenStr, ok := strings.CutPrefix(authorization, "Bearer ")
	if !ok || tokenStr == "" {
		return "", false
	}

	token, err := jwtlib.ParseWithClaims(
		tokenStr,
		&jwtlib.RegisteredClaims{},
		func(t *jwtlib.Token) (any, error) {
			return []byte(m.env.JwtSecret), nil
		},
	)
	if err != nil || !token.Valid {
		return "", false
	}

	claims, ok := token.Claims.(*jwtlib.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return "", false
	}

	return claims.Subject, true
}
