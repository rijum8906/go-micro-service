// Package metadata provides gRPC metadata helpers
package metadata

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const (
	ClientInfoKey contextKey = "client_info"
	AuthzUserKey  contextKey = "authz_user"
)

// Metadata keys (gRPC headers)
const (
	// Client info
	metaKeyDeviceID   = "device-id"
	metaKeyUserAgent  = "user-agent"
	metaKeyClientIP   = "client-ip"
	metaKeyClientType = "client-type"

	// Versioning
	metaKeyAPIVersion = "api-version" // v1, v2, etc.
	metaKeySDKVersion = "sdk-version" // 1.0.0, 2.1.3
	// metaKeyAppVersion = "app-version" // 2.1.0

	// Platform
	metaKeyPlatform = "platform" // ios, android, web
	// metaKeyPlatformVer = "platform-version" // 17.0, 14, Chrome 120

	// Tracing
	metaKeyRequestID = "request-id"
	// metaKeyTraceID   = "trace-id"
	// metaKeySpanID    = "span-id"
	// metaKeySessionID = "session-id"

	// Environment
	// metaKeyLocale   = "locale"   // en-US
	// metaKeyTimezone = "timezone" // America/New_York

	// Auth
	metaKeyUserID       = "user-id"
	metaKeyAccessToken  = "access-token"
	metaKeyRefreshToken = "refresh-token"
)

// ClientInfo represents comprehensive client request metadata
type ClientInfo struct {
	// Core
	DeviceID   string
	UserAgent  string
	IPAddress  string
	ClientType string

	// Versioning
	APIVersion string // API version requested
	SDKVersion string // Client SDK version
	// AppVersion string // App version

	// Platform
	Platform string // ios, android, web, desktop
	// PlatformVer string // OS or browser version

	// Tracing
	RequestID string // Unique request identifier
	// TraceID   string // Distributed trace ID
	// SpanID    string // Current span ID
	// SessionID string // User session ID

	// Environment
	// Locale   string // Language/region
	// Timezone string // Client timezone
}

// AuthzUserInfo represents authenticated user metadata
type AuthzUserInfo struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

// SendClientInfo adds client info to outgoing gRPC metadata
func SendClientInfo(ctx context.Context, info ClientInfo) context.Context {
	md := metadata.Pairs(
		// Core
		metaKeyDeviceID, info.DeviceID,
		metaKeyUserAgent, info.UserAgent,
		metaKeyClientIP, info.IPAddress,
		metaKeyClientType, info.ClientType,

		// Versioning
		metaKeyAPIVersion, info.APIVersion,
		metaKeySDKVersion, info.SDKVersion,
		// metaKeyAppVersion, info.AppVersion,

		// Platform
		metaKeyPlatform, info.Platform,
		// metaKeyPlatformVer, info.PlatformVer,

		// Tracing
		metaKeyRequestID, info.RequestID,
		// metaKeyTraceID, info.TraceID,
		// metaKeySpanID, info.SpanID,
		// metaKeySessionID, info.SessionID,

		// Environment
		// metaKeyLocale, info.Locale,
		// metaKeyTimezone, info.Timezone,
	)

	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existing, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

// ReceiveClientInfo extracts client info from incoming gRPC metadata
func ReceiveClientInfo(ctx context.Context) (ClientInfo, *apperror.AppError) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ClientInfo{}, apperror.ErrInternal.WithMessage("No metadata found in context")
	}

	return ClientInfo{
		// Core
		DeviceID:   get(md, metaKeyDeviceID),
		UserAgent:  get(md, metaKeyUserAgent),
		IPAddress:  get(md, metaKeyClientIP),
		ClientType: get(md, metaKeyClientType),

		// Versioning
		APIVersion: get(md, metaKeyAPIVersion),
		SDKVersion: get(md, metaKeySDKVersion),
		// AppVersion: get(md, metaKeyAppVersion),

		// Platform
		Platform: get(md, metaKeyPlatform),
		// PlatformVer: get(md, metaKeyPlatformVer),

		// Tracing
		RequestID: get(md, metaKeyRequestID),
		// TraceID:   get(md, metaKeyTraceID),
		// SpanID:    get(md, metaKeySpanID),
		// SessionID: get(md, metaKeySessionID),

		// Environment
		// Locale:   get(md, metaKeyLocale),
		// Timezone: get(md, metaKeyTimezone),
	}, nil
}

// SendAuthzUserInfo adds authorized user info to outgoing gRPC metadata
func SendAuthzUserInfo(ctx context.Context, info AuthzUserInfo) context.Context {
	md := metadata.Pairs(
		metaKeyUserID, info.UserID,
		metaKeyAccessToken, info.AccessToken,
		metaKeyRefreshToken, info.RefreshToken,
	)

	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existing, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

// ReceiveAuthzUserInfo extracts authorized user info from incoming gRPC metadata
func ReceiveAuthzUserInfo(ctx context.Context) (AuthzUserInfo, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return AuthzUserInfo{}, apperror.ErrInternal.WithMessage("No metadata found in context")
	}

	return AuthzUserInfo{
		UserID:       get(md, metaKeyUserID),
		AccessToken:  get(md, metaKeyAccessToken),
		RefreshToken: get(md, metaKeyRefreshToken),
	}, nil
}

func get(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
