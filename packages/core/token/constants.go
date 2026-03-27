package token

type TokenScope string

const (
	ScopeAuth           TokenScope = "auth"
	ScopeResetPassword  TokenScope = "reset_password" // For password resets, etc.
	ScopeChangeEmail    TokenScope = "change_email"
	ScopeChangePassword TokenScope = "change_password"
	ScopeDeleteAccount  TokenScope = "delete_account"
	ScopeDeleteProfile  TokenScope = "delete_profile"
)

type AuthMethod string

const (
	AuthPassword  AuthMethod = "password"
	AuthBiometric AuthMethod = "biometric"
	AuthOTP       AuthMethod = "otp"
	AuthRecovery  AuthMethod = "recovery"
)
