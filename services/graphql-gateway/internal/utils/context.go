package utils

import (
	"context"
	"net/textproto"
	"strings"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/dto/coredto"
)

func getHeadersFromContext(ctx context.Context) map[string][]string {
	headers, ok := ctx.Value("headers").(map[string][]string)
	if !ok {
		return nil
	}

	return headers
}

func getHeaderValues(headers map[string][]string, key string) []string {
	if headers == nil {
		return nil
	}

	if vals := headers[key]; len(vals) > 0 {
		return vals
	}

	canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
	if vals := headers[canonicalKey]; len(vals) > 0 {
		return vals
	}

	lowerKey := strings.ToLower(key)
	if vals := headers[lowerKey]; len(vals) > 0 {
		return vals
	}

	return nil
}

func GetBrowserInfo(ctx context.Context) coredto.BrowserInfo {
	userAgent := "unknown"
	ipAddr := "unknown"

	headers := getHeadersFromContext(ctx)
	if headers == nil {
		return coredto.BrowserInfo{
			UserAgent: userAgent,
			IPAddr:    ipAddr,
		}
	}

	// User Agent
	if vals := getHeaderValues(headers, "User-Agent"); len(vals) > 0 {
		userAgent = vals[0]
	}

	// IP Addr
	if vals := getHeaderValues(headers, "X-Forwarded-For"); len(vals) > 0 {
		ipAddr = vals[0]
	}

	return coredto.BrowserInfo{
		UserAgent: userAgent,
		IPAddr:    ipAddr,
	}
}

func WithBrowserInfo(ctx context.Context, browserInfo coredto.BrowserInfo) context.Context {
	return context.WithValue(ctx, "browserInfo", browserInfo)
}

// GetAccessTokenFromHeader extracts Bearer token from Authorization header
func GetAccessTokenFromHeader(ctx context.Context) (string, *apperror.AppError) {
	headers := getHeadersFromContext(ctx)

	// Get Authorization header
	authHeaders := getHeaderValues(headers, "Authorization")
	if len(authHeaders) == 0 {
		return "", nil // No token present, not an error
	}

	authHeader := authHeaders[0]

	// Check Bearer prefix
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", apperror.New(apperror.CodeUnAuthenticated, "invalid authorization format, expected Bearer token")
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", apperror.New(apperror.CodeUnAuthenticated, "empty token")
	}

	return token, nil
}
