package metadata_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/stretchr/testify/require"
)

func TestSetAndGetUserInfo(t *testing.T) {
	ctx := context.Background()
	userInfo := dto.UserInfo{
		UserID:    uuid.NewString(),
		TokenID:   uuid.NewString(),
		SessionID: uuid.NewString(),
	}

	// Set user info in context
	ctx = metadata.SetUserInfoToOutgoingContext(ctx, userInfo)

	// Get user info from context
	userInfo, ok := metadata.GetUserInfoFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, userInfo.UserID, userInfo.UserID)
	require.Equal(t, userInfo.TokenID, userInfo.TokenID)
	require.Equal(t, userInfo.SessionID, userInfo.SessionID)
}

func TestSetAndGetClientInfo(t *testing.T) {
	ctx := context.Background()
	clientInfo := dto.ClientInfo{
		DeviceID:   uuid.NewString(),
		UserAgent:  "test",
		IPAddress:  "127.0.0.1",
		ClientType: "test",
		APIVersion: "v1",
		Locale:     "en",
		SDKVersion: "v1",
		RequestID:  uuid.NewString(),
		TraceID:    uuid.NewString(),
	}

	// Set client info in context
	ctx = metadata.SetClientInfoToOutgoingContext(ctx, clientInfo)

	// Get client info from context
	clientInfo, ok := metadata.GetClientInfoFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, clientInfo.DeviceID, clientInfo.DeviceID)
	require.Equal(t, clientInfo.UserAgent, clientInfo.UserAgent)
	require.Equal(t, clientInfo.IPAddress, clientInfo.IPAddress)
	require.Equal(t, clientInfo.ClientType, clientInfo.ClientType)
	require.Equal(t, clientInfo.APIVersion, clientInfo.APIVersion)
	require.Equal(t, clientInfo.Locale, clientInfo.Locale)
	require.Equal(t, clientInfo.SDKVersion, clientInfo.SDKVersion)
	require.Equal(t, clientInfo.RequestID, clientInfo.RequestID)
	require.Equal(t, clientInfo.TraceID, clientInfo.TraceID)
}

func TestSetAndGetScopedTokenInfo(t *testing.T) {
	ctx := context.Background()
	scopedToken := dto.ScopedToken{
		String:  uuid.NewString(),
		ID:      uuid.NewString(),
		Scope:   "test",
		Subject: "test",
	}

	// Set scoped token info in context
	ctx = metadata.SetScopedTokenInfoToOutgoingContext(ctx, scopedToken)

	// Get scoped token info from context
	scopedToken, ok := metadata.GetScopedTokenInfoFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, scopedToken.String, scopedToken.String)
	require.Equal(t, scopedToken.ID, scopedToken.ID)
	require.Equal(t, scopedToken.Scope, scopedToken.Scope)
	require.Equal(t, scopedToken.Subject, scopedToken.Subject)
}

func TestSetAndGetAuthTokensInfo(t *testing.T) {
	ctx := context.Background()
	authTokens := dto.AuthTokens{
		AccessToken:  uuid.NewString(),
		RefreshToken: uuid.NewString(),
	}

	// Set auth tokens info in context
	ctx = metadata.SetAuthTokensInfoToOutgoingContext(ctx, authTokens)

	// Get auth tokens info from context
	authTokens, ok := metadata.GetAuthTokensInfoFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, authTokens.AccessToken, authTokens.AccessToken)
	require.Equal(t, authTokens.RefreshToken, authTokens.RefreshToken)
}
