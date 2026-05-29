// Package constants is a collection of constants used in the application
package constants

import "slices"

const (
	// Core
	TokenScopeRefreshToken = "USER_REFRESH_TOKEN"

	// Email
	TokenScopeVerifyEmail = "VERIFY_USER_EMAIL"

	// Password
	TokenScopeChangePassword = "CHANGE_USER_PASSWORD"
	TokenScopeResetPassword  = "RESET_USER_PASSWORD"

	// 2FA
	TokenScope2FA = "TWO_FACTOR_AUTHENTICATION"
)

var TokenScopes = []string{
	TokenScopeRefreshToken,
	TokenScopeVerifyEmail,
	TokenScopeChangePassword,
	TokenScopeResetPassword,
	TokenScope2FA,
}

func IsValidaTokenScope(tokenScope string) bool {
	return slices.Contains(TokenScopes, tokenScope)
}
