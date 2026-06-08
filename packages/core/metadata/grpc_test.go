package metadata_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetUserInfoToOutgoingContext(t *testing.T) {
	ctx := context.Background()
	userInfo := dto.UserInfo{
		UserID:    uuid.NewString(),
		TokenID:   uuid.NewString(),
		SessionID: uuid.NewString(),
	}

	// Set user info in context
	ctx = metadata.SetUserInfoToOutgoingContext(ctx, userInfo)

	// Get user info from context
	receivedInfo, ok := testutils.GetUserInfoFromOutgoingContext(ctx)
	if !ok {
		t.Fatal("failed to get metadata from outgoing context")
	}
	assert.Equal(t, userInfo.UserID, receivedInfo.UserID)
	assert.Equal(t, userInfo.TokenID, receivedInfo.TokenID)
	assert.Equal(t, userInfo.SessionID, receivedInfo.SessionID)
}

func TestGetUserFromIncomingContext(t *testing.T) {
	ctx := context.Background()
	userInfo := dto.UserInfo{
		UserID:    uuid.NewString(),
		TokenID:   uuid.NewString(),
		SessionID: uuid.NewString(),
	}

	// Set user info in context
	ctx = testutils.SetUserInfoToIncomingContext(ctx, userInfo)

	// Get user info from context
	receivedInfo, ok := metadata.GetUserInfoFromIncomingContext(ctx)
	if !ok {
		t.Fatal("failed to get metadata from incoming context")
	}
	assert.Equal(t, userInfo.UserID, receivedInfo.UserID)
	assert.Equal(t, userInfo.TokenID, receivedInfo.TokenID)
	assert.Equal(t, userInfo.SessionID, receivedInfo.SessionID)
}

func TestSetClientInfoToOutgoingContext(t *testing.T) {
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
	receivedInfo, ok := testutils.GetClientInfoFromOutgoingContext(ctx)
	if !ok {
		t.Fatal("failed to get metadata from outgoing context")
	}
	require.Equal(t, receivedInfo.DeviceID, clientInfo.DeviceID)
	require.Equal(t, receivedInfo.UserAgent, clientInfo.UserAgent)
	require.Equal(t, receivedInfo.IPAddress, clientInfo.IPAddress)
	require.Equal(t, receivedInfo.ClientType, clientInfo.ClientType)
	require.Equal(t, receivedInfo.APIVersion, clientInfo.APIVersion)
	require.Equal(t, receivedInfo.Locale, clientInfo.Locale)
	require.Equal(t, receivedInfo.SDKVersion, clientInfo.SDKVersion)
	require.Equal(t, receivedInfo.RequestID, clientInfo.RequestID)
	require.Equal(t, receivedInfo.TraceID, clientInfo.TraceID)
}

func TestGetClientInfoFromIncomingContext(t *testing.T) {
	ctx := context.Background()
	clientInfo := dto.ClientInfo{
		DeviceID:   uuid.NewString(),
		UserAgent:  "test",
		IPAddress:  "test",
		ClientType: "test",
		APIVersion: "test",
		Locale:     "test",
		SDKVersion: "test",
		RequestID:  uuid.NewString(),
		TraceID:    uuid.NewString(),
	}

	// Set client info to incoming context
	ctx = testutils.SetClientInfoToIncomingContext(ctx, clientInfo)

	// Get client info from context
	receivedInfo, ok := metadata.GetClientInfoFromIncomingContext(ctx)
	if !ok {
		t.Fatal("failed to get client info from incoming context")
	}
	require.Equal(t, receivedInfo.DeviceID, clientInfo.DeviceID)
	require.Equal(t, receivedInfo.UserAgent, clientInfo.UserAgent)
	require.Equal(t, receivedInfo.IPAddress, clientInfo.IPAddress)
	require.Equal(t, receivedInfo.ClientType, clientInfo.ClientType)
	require.Equal(t, receivedInfo.APIVersion, clientInfo.APIVersion)
	require.Equal(t, receivedInfo.Locale, clientInfo.Locale)
	require.Equal(t, receivedInfo.SDKVersion, clientInfo.SDKVersion)
	require.Equal(t, receivedInfo.RequestID, clientInfo.RequestID)
	require.Equal(t, receivedInfo.TraceID, clientInfo.TraceID)
}

func TestSetScopedTokenInfoToOutgoingContext(t *testing.T) {
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
	receivedInfo, ok := testutils.GetScopedTokenInfoFromOutgoingContext(ctx)
	if !ok {
		t.Fatal("failed to get metadata from outgoing context")
	}
	require.Equal(t, receivedInfo.String, scopedToken.String)
	require.Equal(t, receivedInfo.ID, scopedToken.ID)
	require.Equal(t, receivedInfo.Scope, scopedToken.Scope)
	require.Equal(t, receivedInfo.Subject, scopedToken.Subject)
}

func TestGetScopedTokenInfoFromIncomingContext(t *testing.T) {
	ctx := context.Background()
	scopedToken := dto.ScopedToken{
		String:  uuid.NewString(),
		ID:      uuid.NewString(),
		Scope:   "test",
		Subject: "test",
	}

	// Set scoped token info in context
	ctx = testutils.SetScopedTokenInfoToIncomingContext(ctx, scopedToken)

	// Get scoped token info from context
	receivedInfo, ok := metadata.GetScopedTokenInfoFromIncomingContext(ctx)
	if !ok {
		t.Fatal("failed to get metadata from incoming context")
	}
	require.Equal(t, receivedInfo.String, scopedToken.String)
	require.Equal(t, receivedInfo.ID, scopedToken.ID)
	require.Equal(t, receivedInfo.Scope, scopedToken.Scope)
	require.Equal(t, receivedInfo.Subject, scopedToken.Subject)
}

func TestSetAuthTokensInfoToOutgoingContext(t *testing.T) {
	ctx := context.Background()
	authTokens := dto.AuthTokens{
		AccessToken:  uuid.NewString(),
		RefreshToken: uuid.NewString(),
	}

	// Set auth tokens info in context
	ctx = metadata.SetAuthTokensInfoToOutgoingContext(ctx, authTokens)

	// Get auth tokens info from context
	receivedInfo, ok := testutils.GetAuthTokensInfoFromOutgoingContext(ctx)
	if !ok {
		t.Fatal("failed to get metadata from outgoing context")
	}
	require.Equal(t, receivedInfo.AccessToken, authTokens.AccessToken)
	require.Equal(t, receivedInfo.RefreshToken, authTokens.RefreshToken)
}

func TestGetAuthTokensInfoFromIncomingContext(t *testing.T) {
	ctx := context.Background()
	authTokens := dto.AuthTokens{
		AccessToken:  uuid.NewString(),
		RefreshToken: uuid.NewString(),
	}

	// Set auth tokens info in context
	ctx = testutils.SetAuthTokensInfoToIncomingContext(ctx, authTokens)

	// Get auth tokens info from context
	receivedInfo, ok := metadata.GetAuthTokensInfoFromIncomingContext(ctx)
	if !ok {
		t.Fatal("failed to get metadata from incoming context")
	}
	require.Equal(t, receivedInfo.AccessToken, authTokens.AccessToken)
	require.Equal(t, receivedInfo.RefreshToken, authTokens.RefreshToken)
}
