// Package metadata provides utility functions for working with gRPC metadata.
package metadata

import (
	"context"

	"github.com/rijum8906/relay/packages/core/dto"
	"google.golang.org/grpc/metadata"
)

func NewOutgoingContext(ctx context.Context, md metadata.MD) context.Context {
	return metadata.NewOutgoingContext(ctx, md)
}

func WithOutgoingContext(ctx context.Context, key string, values ...string) (context.Context, bool) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return nil, false
	}

	md.Set(key, values...)

	return metadata.NewOutgoingContext(ctx, md), true
}

func SetUserInfoToOutgoingContext(ctx context.Context, userInfo dto.UserInfo) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	// Set user info to metadata
	md.Set(dto.MetaUserIDKey, userInfo.UserID)
	md.Set(dto.MetaTokenIDKey, userInfo.TokenID)
	md.Set(dto.MetaSessionIDKey, userInfo.SessionID)

	return metadata.NewOutgoingContext(ctx, md)
}

func GetUserInfoFromContext(ctx context.Context) (dto.UserInfo, bool) {
	info := dto.UserInfo{}
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return info, false
	}

	if userID, ok := md[dto.MetaUserIDKey]; ok {
		info.UserID = userID[0]
	}
	if tokenID, ok := md[dto.MetaTokenIDKey]; ok {
		info.TokenID = tokenID[0]
	}
	if sessionID, ok := md[dto.MetaSessionIDKey]; ok {
		info.SessionID = sessionID[0]
	}

	return info, true
}

func SetClientInfoToOutgoingContext(ctx context.Context, clientInfo dto.ClientInfo) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	md.Set(dto.MetaDeviceIDKey, clientInfo.DeviceID)
	md.Set(dto.MetaUserAgentKey, clientInfo.UserAgent)
	md.Set(dto.MetaClientIPKey, clientInfo.IPAddress)
	md.Set(dto.MetaClientTypeKey, clientInfo.ClientType)
	md.Set(dto.MetaAPIVersionKey, clientInfo.APIVersion)
	md.Set(dto.MetaLocaleKey, clientInfo.Locale)
	md.Set(dto.MetaSDKVersionKey, clientInfo.SDKVersion)

	md.Set(dto.MetaRequestIDKey, clientInfo.RequestID)
	md.Set(dto.MetaTraceIDKey, clientInfo.TraceID)

	return metadata.NewOutgoingContext(ctx, md)
}

func GetClientInfoFromContext(ctx context.Context) (dto.ClientInfo, bool) {
	info := dto.ClientInfo{}
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return info, false
	}

	if deviceID, ok := md[dto.MetaDeviceIDKey]; ok {
		info.DeviceID = deviceID[0]
	}
	if userAgent, ok := md[dto.MetaUserAgentKey]; ok {
		info.UserAgent = userAgent[0]
	}
	if clientIP, ok := md[dto.MetaClientIPKey]; ok {
		info.IPAddress = clientIP[0]
	}
	if clientType, ok := md[dto.MetaClientTypeKey]; ok {
		info.ClientType = clientType[0]
	}
	if apiVersion, ok := md[dto.MetaAPIVersionKey]; ok {
		info.APIVersion = apiVersion[0]
	}
	if locale, ok := md[dto.MetaLocaleKey]; ok {
		info.Locale = locale[0]
	}
	if sdkVersion, ok := md[dto.MetaSDKVersionKey]; ok {
		info.SDKVersion = sdkVersion[0]
	}

	if requestID, ok := md[dto.MetaRequestIDKey]; ok {
		info.RequestID = requestID[0]
	}
	if traceID, ok := md[dto.MetaTraceIDKey]; ok {
		info.TraceID = traceID[0]
	}

	return info, true
}

func SetScopedTokenInfoToOutgoingContext(ctx context.Context, tokenInfo dto.ScopedToken) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	md.Set(dto.MetaScopedTokenKey, tokenInfo.String)
	md.Set(dto.MetaScopedTokenIDKey, tokenInfo.ID)
	md.Set(dto.MetaScopedTokenScopeKey, tokenInfo.Scope)
	md.Set(dto.MetaScopedTokenSubjectKey, tokenInfo.Subject)

	return metadata.NewOutgoingContext(ctx, md)
}

func GetScopedTokenInfoFromContext(ctx context.Context) (dto.ScopedToken, bool) {
	info := dto.ScopedToken{}
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return info, false
	}

	if scopedToken, ok := md[dto.MetaScopedTokenKey]; ok {
		info.String = scopedToken[0]
	}
	if scopedTokenID, ok := md[dto.MetaScopedTokenIDKey]; ok {
		info.ID = scopedTokenID[0]
	}
	if scopedTokenScope, ok := md[dto.MetaScopedTokenScopeKey]; ok {
		info.Scope = scopedTokenScope[0]
	}
	if scopedTokenSubject, ok := md[dto.MetaScopedTokenSubjectKey]; ok {
		info.Subject = scopedTokenSubject[0]
	}

	return info, true
}

func SetAuthTokensInfoToOutgoingContext(ctx context.Context, tokenInfo dto.AuthTokens) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	md.Set(dto.MetaAccessTokenKey, tokenInfo.AccessToken)
	md.Set(dto.MetaRefreshTokenKey, tokenInfo.RefreshToken)

	return metadata.NewOutgoingContext(ctx, md)
}

func GetAuthTokensInfoFromContext(ctx context.Context) (dto.AuthTokens, bool) {
	info := dto.AuthTokens{}
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return info, false
	}

	if accessToken, ok := md[dto.MetaAccessTokenKey]; ok {
		info.AccessToken = accessToken[0]
	}
	if refreshToken, ok := md[dto.MetaRefreshTokenKey]; ok {
		info.RefreshToken = refreshToken[0]
	}

	return info, true
}
