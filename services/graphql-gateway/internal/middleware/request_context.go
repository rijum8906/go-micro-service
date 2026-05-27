package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/metadata"
)

func ExtractClientInfo(r *http.Request) dto.ClientInfo {
	return dto.ClientInfo{
		DeviceID:   r.Header.Get(dto.DeviceIDReqHeaderKey),
		UserAgent:  r.Header.Get(dto.UserAgentReqHeaderKey),
		IPAddress:  r.Header.Get(dto.ClientIPReqHeaderKey),
		ClientType: r.Header.Get(dto.ClientTypeReqHeaderKey),
		APIVersion: r.Header.Get(dto.APIVersionReqHeaderKey),
		SDKVersion: r.Header.Get(dto.SDKVersionReqHeaderKey),
		RequestID:  uuid.NewString(),
		TraceID:    uuid.NewString(),
		Locale:     r.Header.Get(dto.LocaleReqHeaderKey),
	}
}

func ExtractAuthTokens(r *http.Request) dto.AuthTokens {
	accessToken := r.Header.Get(dto.AuthorizationReqHeaderKey)
	if accessToken != "" {
		tokenParts := strings.Split(accessToken, " ")
		if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
			accessToken = tokenParts[1]
		}
	} else {
		// Fallback to cookie
		res, err := r.Cookie(dto.MetaAccessTokenKey)
		if err == nil {
			accessToken = res.Value
		}
	}

	refreshToken := r.Header.Get(dto.RefreshTokenReqHeaderKey)
	if refreshToken == "" {
		// Fallback to cookie
		res, err := r.Cookie(dto.MetaRefreshTokenKey)
		if err == nil {
			refreshToken = res.Value
		}
	}

	return dto.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func WithRequestHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientInfo := ExtractClientInfo(r)
		tokens := ExtractAuthTokens(r)

		ctx := r.Context()
		ctx = metadata.SendClientInfo(ctx, clientInfo)
		ctx = metadata.SendTokensInfo(ctx, tokens)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
