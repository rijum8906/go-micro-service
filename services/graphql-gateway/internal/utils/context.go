package utils

import (
	"context"

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
