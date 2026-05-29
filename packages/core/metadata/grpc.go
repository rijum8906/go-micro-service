// Package metadata provides helpers for reading and writing gRPC metadata.
package metadata

import (
	"context"

	"github.com/rijum8906/relay/packages/core/dto"
	"google.golang.org/grpc/metadata"
)

// Send adds client info to outgoing gRPC metadata.
func Send(ctx context.Context, info dto.ClientInfo) context.Context {
	return SendClientInfo(ctx, info)
}

// Receive extracts client info from incoming gRPC metadata.
func Receive(ctx context.Context) (dto.ClientInfo, bool) {
	return ReceiveClientInfo(ctx)
}

// SendClientInfo adds client info to outgoing gRPC metadata.
func SendClientInfo(ctx context.Context, info dto.ClientInfo) context.Context {
	return appendOutgoing(
		ctx,
		dto.MetaDeviceIDKey, info.DeviceID,
		dto.MetaUserAgentKey, info.UserAgent,
		dto.MetaClientIPKey, info.IPAddress,
		dto.MetaClientTypeKey, info.ClientType,
		dto.MetaAPIVersionKey, info.APIVersion,
		dto.MetaSDKVersionKey, info.SDKVersion,
		dto.MetaRequestIDKey, info.RequestID,
		dto.MetaTraceIDKey, info.TraceID,
		dto.MetaLocaleKey, info.Locale,
	)
}

// ReceiveClientInfo extracts client info from incoming gRPC metadata.
func ReceiveClientInfo(ctx context.Context) (dto.ClientInfo, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return dto.ClientInfo{}, false
	}

	return dto.ClientInfo{
		DeviceID:   first(md, dto.MetaDeviceIDKey),
		UserAgent:  first(md, dto.MetaUserAgentKey),
		IPAddress:  first(md, dto.MetaClientIPKey),
		ClientType: first(md, dto.MetaClientTypeKey),
		APIVersion: first(md, dto.MetaAPIVersionKey),
		SDKVersion: first(md, dto.MetaSDKVersionKey),
		RequestID:  first(md, dto.MetaRequestIDKey),
		TraceID:    first(md, dto.MetaTraceIDKey),
		Locale:     first(md, dto.MetaLocaleKey),
	}, true
}

// SendUserInfo adds user identity metadata to the outgoing context.
func SendUserInfo(ctx context.Context, info dto.UserInfo) context.Context {
	return appendOutgoing(
		ctx,
		dto.MetaUserIDKey, info.UserID,
		dto.MetaTokenIDKey, info.TokenID,
		dto.MetaSessionIDKey, info.SessionID,
	)
}

// ReceiveUserInfo extracts user identity metadata from the incoming context.
func ReceiveUserInfo(ctx context.Context) (dto.UserInfo, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return dto.UserInfo{}, false
	}

	return dto.UserInfo{
		UserID:    first(md, dto.MetaUserIDKey),
		TokenID:   first(md, dto.MetaTokenIDKey),
		SessionID: first(md, dto.MetaSessionIDKey),
	}, true
}

func SendAuthTokensInfo(ctx context.Context, tokens dto.AuthTokens) context.Context {
	return appendOutgoing(ctx,
		dto.MetaAccessTokenKey, tokens.AccessToken,
		dto.MetaRefreshTokenKey, tokens.RefreshToken,
	)
}

func ReceiveAuthTokensInfo(ctx context.Context) (dto.AuthTokens, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return dto.AuthTokens{}, false
	}

	return dto.AuthTokens{
		AccessToken:  first(md, dto.MetaAccessTokenKey),
		RefreshToken: first(md, dto.MetaRefreshTokenKey),
	}, true
}

func SendScopedTokenInfo(ctx context.Context, info dto.ScopedToken) context.Context {
	return appendOutgoing(ctx,
		dto.MetaScopedTokenKey, info.String,
		dto.MetaScopedTokenIDKey, info.ID,
		dto.MetaScopedTokenScopeKey, info.Scope,
		dto.MetaScopedTokenSubjectKey, info.Subject,
	)
}

func ReceiveScopedTokenInfo(ctx context.Context) (dto.ScopedToken, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return dto.ScopedToken{}, false
	}

	return dto.ScopedToken{
		String:  first(md, dto.MetaScopedTokenKey),
		ID:      first(md, dto.MetaScopedTokenIDKey),
		Scope:   first(md, dto.MetaScopedTokenScopeKey),
		Subject: first(md, dto.MetaScopedTokenSubjectKey),
	}, true
}

func appendOutgoing(ctx context.Context, kv ...string) context.Context {
	md := filterEmptyPairs(kv...)
	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		merged := metadata.MD{}
		for key, values := range existing {
			if len(values) == 0 {
				continue
			}
			merged[key] = append([]string(nil), values...)
		}
		for key, values := range md {
			merged.Set(key, values...)
		}
		md = merged
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
