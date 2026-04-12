package template

type TemplateType string

const (
	TemplateTypeEmailVerification  TemplateType = "email-verification"
	TemplateTypeEmailPasswordReset TemplateType = "email-password-reset"
	TemplateTypeEmailWelcome       TemplateType = "email-welcome"
)

type WelcomeTemplateDTO struct {
	ClientName  string `validate:"required"`
	ClientEmail string `validate:"required,email"`
}

type EmailVerificationDTO struct {
	ClientName        string `validate:"required"`
	ClientEmail       string `validate:"required,email"`
	VerificationToken string `validate:"required"`
	Validity          string `validate:"required"` // "15 minutes" or "1 hour"
}
type PasswordResetDTO struct {
	ClientName  string `validate:"required"`
	ClientEmail string `validate:"required,email"`
	ResetToken  string `validate:"required"`
	Validity    string `validate:"required"` // "15 minutes" or "1 hour"
}
