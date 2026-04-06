package templates

import (
	"embed"
	"path"

	"github.com/rijum8906/relay/packages/core/apperror"
)

const (
	BaseDir      = "email"
	SubjectFile  = "subject.tmpl"
	TextBodyFile = "body.txt.tmpl"
	HTMLBodyFile = "body.html.tmpl"
)

type EmailTemplate string

const (
	EmailTemplateWelcome       EmailTemplate = "welcome"
	EmailTemplateResetPassword EmailTemplate = "reset-password"
	EmailTemplateVerifyEmail   EmailTemplate = "verify-email"
)

var emailTemplates = []EmailTemplate{
	EmailTemplateWelcome,
	EmailTemplateResetPassword,
	EmailTemplateVerifyEmail,
}

var emailTemplateSet = map[EmailTemplate]struct{}{
	EmailTemplateWelcome:       {},
	EmailTemplateResetPassword: {},
	EmailTemplateVerifyEmail:   {},
}

// FS embeds the canonical on-disk email template tree.
//
// Expected structure:
//   - email/<name>/subject.tmpl
//   - email/<name>/body.txt.tmpl
//   - email/<name>/body.html.tmpl
//
//go:embed email
var FS embed.FS

func (t EmailTemplate) String() string {
	return string(t)
}

func (t EmailTemplate) Validate() *apperror.AppError {
	if t == "" {
		return validationError("template", "template is required")
	}

	if !Exists(t) {
		return validationError("template", "invalid email template")
	}

	return nil
}

func Names() []EmailTemplate {
	return append([]EmailTemplate(nil), emailTemplates...)
}

func Exists(name EmailTemplate) bool {
	_, ok := emailTemplateSet[name]
	return ok
}

func Dir(name EmailTemplate) string {
	return path.Join(BaseDir, name.String())
}

func File(name EmailTemplate, file string) string {
	return path.Join(Dir(name), file)
}

func Files(name EmailTemplate) []string {
	return []string{
		File(name, SubjectFile),
		File(name, TextBodyFile),
		File(name, HTMLBodyFile),
	}
}

func validationError(field, message string) *apperror.AppError {
	return apperror.New(apperror.CodeValidation, message).WithDetail("field", field)
}
