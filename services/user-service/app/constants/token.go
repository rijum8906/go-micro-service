// Package constants is a collection of constants used in the application
package constants

import "slices"

const (
	// Password
	TokenScopeChangePassword = "CHANGE_USER_PASSWORD"

	// Internal (Can't be issued by user)
	TokenScopeResetPassword = "RESET_USER_PASSWORD"
	TokenScopeVerifyEmail   = "VERIFY_USER_EMAIL"
	TokenScopeRefreshToken  = "USER_REFRESH_TOKEN"
	TokenScope2FA           = "TWO_FACTOR_AUTHENTICATION"
)

var TokenScopesInternal = []string{
	TokenScopeResetPassword,
	TokenScopeVerifyEmail,
	TokenScopeRefreshToken,
	TokenScope2FA,
}

func IsInternalTokenScope(tokenScope string) bool {
	return slices.Contains(TokenScopesInternal, tokenScope)
}
