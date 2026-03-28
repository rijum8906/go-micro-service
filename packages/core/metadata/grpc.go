// Package metadata provides helpers for reading and writing gRPC metadata.
package metadata

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	MetaDeviceIDKey     = "device-id"
	MetaUserAgentKey    = "user-agent"
	MetaClientIPKey     = "client-ip"
	MetaClientTypeKey   = "client-type"
	MetaAPIVersionKey   = "api-version"
	MetaSDKVersionKey   = "sdk-version"
	MetaRequestIDKey    = "request-id"
	MetaSessionIDKey    = "session-id"
	MetaTraceIDKey      = "trace-id"
	MetaLocaleKey       = "locale"
	MetaUserIDKey       = "user-id"
	MetaAccessTokenKey  = "access-token"
	MetaRefreshTokenKey = "refresh-token"
)

type UserInfo struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

type ClientInfo struct {
	DeviceID   string
	UserAgent  string
	IPAddress  string
	ClientType string
	APIVersion string
	SDKVersion string
	RequestID  string
	SessionID  string
	TraceID    string
	Locale     string
}

// Send adds client info to outgoing gRPC metadata.
func Send(ctx context.Context, info ClientInfo) context.Context {
	return SendClientInfo(ctx, info)
}

// Receive extracts client info from incoming gRPC metadata.
func Receive(ctx context.Context) (ClientInfo, bool) {
	return ReceiveClientInfo(ctx)
}

// SendClientInfo adds client info to outgoing gRPC metadata.
func SendClientInfo(ctx context.Context, info ClientInfo) context.Context {
	return appendOutgoing(ctx,
		MetaDeviceIDKey, info.DeviceID,
		MetaUserAgentKey, info.UserAgent,
		MetaClientIPKey, info.IPAddress,
		MetaClientTypeKey, info.ClientType,
		MetaAPIVersionKey, info.APIVersion,
		MetaSDKVersionKey, info.SDKVersion,
		MetaRequestIDKey, info.RequestID,
		MetaSessionIDKey, info.SessionID,
		MetaTraceIDKey, info.TraceID,
		MetaLocaleKey, info.Locale,
	)
}

// SendClinetInfo preserves backward compatibility with the previous typoed API.
func SendClinetInfo(ctx context.Context, info ClientInfo) context.Context {
	return SendClientInfo(ctx, info)
}

// ReceiveClientInfo extracts client info from incoming gRPC metadata.
func ReceiveClientInfo(ctx context.Context) (ClientInfo, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ClientInfo{}, false
	}

	return ClientInfo{
		DeviceID:   first(md, MetaDeviceIDKey),
		UserAgent:  first(md, MetaUserAgentKey),
		IPAddress:  first(md, MetaClientIPKey),
		ClientType: first(md, MetaClientTypeKey),
		APIVersion: first(md, MetaAPIVersionKey),
		SDKVersion: first(md, MetaSDKVersionKey),
		RequestID:  first(md, MetaRequestIDKey),
		SessionID:  first(md, MetaSessionIDKey),
		TraceID:    first(md, MetaTraceIDKey),
		Locale:     first(md, MetaLocaleKey),
	}, true
}

// SendUserInfo adds user identity metadata to the outgoing context.
func SendUserInfo(ctx context.Context, info UserInfo) context.Context {
	return appendOutgoing(ctx,
		MetaUserIDKey, info.UserID,
		MetaAccessTokenKey, info.AccessToken,
		MetaRefreshTokenKey, info.RefreshToken,
	)
}

// ReceiveUserInfo extracts user identity metadata from the incoming context.
func ReceiveUserInfo(ctx context.Context) (UserInfo, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return UserInfo{}, false
	}

	return UserInfo{
		UserID:       first(md, MetaUserIDKey),
		AccessToken:  first(md, MetaAccessTokenKey),
		RefreshToken: first(md, MetaRefreshTokenKey),
	}, true
}

func appendOutgoing(ctx context.Context, kv ...string) context.Context {
	md := filterEmptyPairs(kv...)
	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existing, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

func filterEmptyPairs(kv ...string) metadata.MD {
	md := metadata.MD{}
	for i := 0; i+1 < len(kv); i += 2 {
		key := kv[i]
		value := kv[i+1]
		if key == "" || value == "" {
			continue
		}
		md.Append(key, value)
	}

	return md
}

func first(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
