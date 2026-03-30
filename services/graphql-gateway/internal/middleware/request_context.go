package middleware

import (
	"context"
	"net"
	"net/http"
)

func WithRequestHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := make(map[string][]string, len(r.Header)+1)
		for key, vals := range r.Header {
			copied := append([]string(nil), vals...)
			headers[key] = copied
		}

		if len(headers["X-Forwarded-For"]) == 0 && r.RemoteAddr != "" {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil && host != "" {
				headers["X-Forwarded-For"] = []string{host}
			}
		}

		ctx := context.WithValue(r.Context(), "headers", headers)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
