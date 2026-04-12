package template

type TemplateType string

const (
	TemplateTypeEmailVerification  TemplateType = "email-verification"
	TemplateTypeEmailPasswordReset TemplateType = "email-password-reset"
	TemplateTypeEmailWelcome       TemplateType = "email-welcome"
)

type WelcomeTemplateDTO struct {
	ClientName  string
	ClientEmail string
}

type EmailVerificationDTO struct {
	ClientName        string
	ClientEmail       string
	VerificationToken string
	Validity          string // eg. "10 minutes" or "1 hour"
}
type PasswordResetDTO struct {
	ClientName  string
	ClientEmail string
	ResetToken  string
	Validity    string
}
