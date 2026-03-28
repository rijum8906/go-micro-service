package utils

import (
	"context"
	"strings"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/dto/coredto"
)

func GetBrowserInfo(ctx context.Context) coredto.BrowserInfo {
	userAgent := "unknown"
	ipAddr := "unknown"

	headers, ok := ctx.Value("headers").(map[string][]string)
	if !ok {
		return coredto.BrowserInfo{
			UserAgent: userAgent,
			IPAddr:    ipAddr,
		}
	}

	// User Agent
	if vals := headers["User-Agent"]; len(vals) > 0 {
		userAgent = vals[0]
	}

	// IP Addr
	if vals := headers["X-Forwarded-For"]; len(vals) > 0 {
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
	// Get headers from context (gqlgen stores them as map[string][]string)
	headers, ok := ctx.Value("headers").(map[string][]string)
	if !ok {
		return "", apperror.New(apperror.TypeValidation, apperror.CodeValidation, "headers not found in context")
	}

	// Get Authorization header
	authHeaders, ok := headers["Authorization"]
	if !ok || len(authHeaders) == 0 {
		return "", nil // No token present, not an error
	}

	authHeader := authHeaders[0]

	// Check Bearer prefix
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", apperror.New(apperror.TypeValidation, apperror.CodeUnAuthenticated, "invalid authorization format, expected Bearer token")
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", apperror.New(apperror.TypeValidation, apperror.CodeUnAuthenticated, "empty token")
	}

	return token, nil
}
