package dto

type WelcomeTemplateDTO struct {
	ClientName  string `validate:"required"`
	ClientEmail string `validate:"required,email"`
}

type EmailVerificationDTO struct {
	ClientName      string `validate:"required"`
	ClientEmail     string `validate:"required,email"`
	VerificationURL string `validate:"required"`
	Validity        string `validate:"required"` // "15 minutes" or "1 hour"
}
type PasswordResetDTO struct {
	ClientName  string `validate:"required"`
	ClientEmail string `validate:"required,email"`
	ResetURL    string `validate:"required"`
	Validity    string `validate:"required"` // "15 minutes" or "1 hour"
}
