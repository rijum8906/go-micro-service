// Package coreconstants
package coreconstants

type AuthMethod string

const (
	AuthMethodPassword       AuthMethod = "PASSWORD"
	AuthMethodBiometric      AuthMethod = "BIOMETRIC"
	AuthMethodOtp            AuthMethod = "OTP"
	AuthMethodTotp           AuthMethod = "TOTP"
	AuthMethodRecovery       AuthMethod = "RECOVERY"
	AuthMethodMagicLink      AuthMethod = "MAGIC_LINK"
	AuthMethodSocialGoogle   AuthMethod = "SOCIAL_GOOGLE"
	AuthMethodSocialGithub   AuthMethod = "SOCIAL_GITHUB"
	AuthMethodAPIKey         AuthMethod = "API_KEY"
	AuthMethodServiceAccount AuthMethod = "SERVICE_ACCOUNT"
)
