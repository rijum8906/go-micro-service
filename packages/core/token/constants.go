package token

// NOTE: do no change this without adding the same types in graphql-gateway
// schema (services/graphql-gateway/internal/schema/core/v1/enums.graphqls)

import "slices"

type TokenScope string

const (
	TokenScopeAuth           TokenScope = "AUTH"
	TokenScopeRefresh        TokenScope = "REFRESH"
	TokenScopeChangeEmail    TokenScope = "CHANGE_EMAIL"
	TokenScopeChangePassword TokenScope = "CHANGE_PASSWORD"
	TokenScopeDeleteAccount  TokenScope = "DELETE_ACCOUNT"
	TokenScopeResetPassword  TokenScope = "RESET_PASSWORD"
	TokenScopeVerifyEmail    TokenScope = "VERIFY_EMAIL"
	TokenScopeEnable2fa      TokenScope = "ENABLE_2FA"
	TokenScopeDisable2fa     TokenScope = "DISABLE_2FA"
	TokenScopeAdmin          TokenScope = "ADMIN"
	TokenScopeImpersonate    TokenScope = "IMPERSONATE"
	TokenScopeRecovery       TokenScope = "RECOVERY"
)

// Organization
const (
	TokenScopeUpdateOrganizationName  TokenScope = "UPDATE_ORG_NAME"
	TokenScopeChangeOrganizationOwner TokenScope = "CHANGE_ORG_OWNER"
)

var tokenScopes = []TokenScope{
	TokenScopeAuth,
	TokenScopeRefresh,
	TokenScopeChangeEmail,
	TokenScopeChangePassword,
	TokenScopeDeleteAccount,
	TokenScopeResetPassword,
	TokenScopeVerifyEmail,
	TokenScopeEnable2fa,
	TokenScopeDisable2fa,
	TokenScopeAdmin,
	TokenScopeImpersonate,
	TokenScopeRecovery,
	TokenScopeUpdateOrganizationName,
	TokenScopeChangeOrganizationOwner,
}

func (t TokenScope) Validate() bool {
	return slices.Contains(tokenScopes, t)
}

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
