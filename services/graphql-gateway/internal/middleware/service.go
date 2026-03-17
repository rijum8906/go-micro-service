package middleware

import "net/http"

func (m *middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			r.Header.Set("X-Is-Authenticated", "false")
		} else {
			r.Header.Set("X-Is-Authenticated", "true")
		}

		ipAddr := r.Header.Get("X-Forwarded-For")
		if ipAddr == "" {
			ipAddr = r.RemoteAddr
		}
		r.Header.Set("X-Client-IP", ipAddr)

		next.ServeHTTP(w, r)
	})
}
