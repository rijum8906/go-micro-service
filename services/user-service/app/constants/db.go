package constants

// TwoFactorMethod represents the method used for two-factor authentication.
// NOTE: update this list when adding new two-factor methods in the database schema.
const (
	TwoFactorMethodTotp     = "totp"
	TwoFactorMethodWebauthn = "webauthn"
	TwoFactorMethodEmail    = "email"
)

func IsValidTwoFactorMethod(method string) bool {
	switch method {
	case TwoFactorMethodTotp, TwoFactorMethodWebauthn, TwoFactorMethodEmail:
		return true
	default:
		return false
	}
}
