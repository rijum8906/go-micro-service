package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountNotFound    = errors.New("account not found")
	ErrProfileNotFound    = errors.New("profile not found")
	ErrOAuthNotFound      = errors.New("oauth not found")
	ErrAccountExists      = errors.New("account already exists")

	// Refresh token
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrExpiredRefreshToken = errors.New("refresh token expired")

	// Token
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token expired")
	ErrInvalidTokenClaims = errors.New("invalid token claims")
	ErrTokenRevoked       = errors.New("token revoked")

	// Session
	ErrInvalidSession  = errors.New("invalid session")
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionNotFound = errors.New("session not found")
)
