package dto

type BaseEmailDTO struct {
	ClientName  string `validate:"required"`
	ClientEmail string `validate:"required,email"`
}

type WelcomeTemplateDTO struct {
	ClientName  string `validate:"required"`
	ClientEmail string `validate:"required,email"`
}

type EmailVerificationDTO struct {
	BaseEmailDTO
	VerificationURL string `validate:"required"`
	Validity        string `validate:"required"` // "15 minutes" or "1 hour"
}
type PasswordResetDTO struct {
	BaseEmailDTO
	ResetURL string `validate:"required"`
	Validity string `validate:"required"` // "15 minutes" or "1 hour"
}
