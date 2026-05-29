package dto

import coreconstants "github.com/rijum8906/relay/packages/core/constants"

// UserInfo holds the user information received from the authentication access token
// Use case: After successful authentication, attach this struct to the request context (by the gateway) to be used by other services
type UserInfo struct {
	UserID    string
	TokenID   string
	SessionID string
}

// ClientInfo holds the client information received from the request headers and some will be added by the server
// Use case: Attach this struct to the request context (by the middleware) to be used by the gateway and other services
// Only for selected service clients' methods
type ClientInfo struct {
	DeviceID   string
	UserAgent  string
	IPAddress  string
	ClientType string
	APIVersion string
	Locale     string
	SDKVersion string

	// Created by the server (gateway)
	RequestID string
	TraceID   string
}

// AuthTokens holds the authentication tokens received from the client
// Use case: Attach this struct to the request context (by the gateway) to be used by other services
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

// ScopedToken holds the scoped token received from the client
// Use case: Attach this struct to the request context (by the gateway) to be used by other services
// Example: Pre auth token for two factor authentication
type ScopedToken struct {
	String  string
	ID      string
	Scope   string
	Subject string
}

// AuthorizationToken holds the authorization token received from the client request header (Authization header)
// Use case: Attach this struct to the request context (by the middleware) to be used by the gateway
type AuthorizationToken struct {
	Token string
	Type  coreconstants.TokenType
}
