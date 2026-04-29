package template

type TemplateType string

const (
	TemplateTypeEmailVerification  TemplateType = "email-verification"
	TemplateTypeEmailPasswordReset TemplateType = "email-password-reset"
	TemplateTypeEmailWelcome       TemplateType = "email-welcome"
)
