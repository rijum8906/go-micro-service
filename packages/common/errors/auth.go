package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountNotFound    = errors.New("account not found")
	ErrProfileNotFound    = errors.New("profile not found")
	ErrOAuthNotFound      = errors.New("oauth not found")

	// Token
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token expired")
	ErrInvalidTokenClaims = errors.New("invalid token claims")
)
