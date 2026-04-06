package templates

import (
	"bytes"
	"path"
	texttemplate "text/template"

	"github.com/rijum8906/relay/packages/core/apperror"
)

type RenderedEmail struct {
	Subject  string `json:"subject"`
	TextBody string `json:"text_body"`
	HTMLBody string `json:"html_body"`
}

func RenderEmail(name EmailTemplate, data map[string]string) (*RenderedEmail, *apperror.AppError) {
	if appErr := name.Validate(); appErr != nil {
		return nil, appErr
	}

	subject, appErr := renderFile(name, SubjectFile, data)
	if appErr != nil {
		return nil, appErr
	}

	textBody, appErr := renderFile(name, TextBodyFile, data)
	if appErr != nil {
		return nil, appErr
	}

	htmlBody, appErr := renderFile(name, HTMLBodyFile, data)
	if appErr != nil {
		return nil, appErr
	}

	return &RenderedEmail{
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}, nil
}

func renderFile(name EmailTemplate, file string, data map[string]string) (string, *apperror.AppError) {
	templatePath := File(name, file)

	tmpl, err := texttemplate.New(path.Base(file)).Option("missingkey=error").ParseFS(FS, templatePath)
	if err != nil {
		return "", apperror.New(apperror.CodeInternal, "failed to parse email template").
			WithDetail("template", name.String()).
			WithDetail("file", file).
			WithDetail("error", err.Error())
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", apperror.New(apperror.CodeValidation, "failed to render email template").
			WithDetail("template", name.String()).
			WithDetail("file", file).
			WithDetail("error", err.Error())
	}

	return out.String(), nil
}
