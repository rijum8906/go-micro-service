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

type RenderedNotification struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	ActionURL string `json:"action_url"`
}

func RenderEmail(name EmailTemplate, data map[string]string) (*RenderedEmail, *apperror.AppError) {
	if appErr := name.Validate(); appErr != nil {
		return nil, appErr
	}

	subject, appErr := renderTemplateFile(File(name, SubjectFile), name.String(), "email", SubjectFile, data)
	if appErr != nil {
		return nil, appErr
	}

	textBody, appErr := renderTemplateFile(File(name, TextBodyFile), name.String(), "email", TextBodyFile, data)
	if appErr != nil {
		return nil, appErr
	}

	htmlBody, appErr := renderTemplateFile(File(name, HTMLBodyFile), name.String(), "email", HTMLBodyFile, data)
	if appErr != nil {
		return nil, appErr
	}

	return &RenderedEmail{
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}, nil
}

func RenderNotification(name NotificationTemplate, data map[string]string) (*RenderedNotification, *apperror.AppError) {
	if appErr := name.Validate(); appErr != nil {
		return nil, appErr
	}

	title, appErr := renderTemplateFile(NotificationFile(name, NotificationTitleFile), name.String(), "notification", NotificationTitleFile, data)
	if appErr != nil {
		return nil, appErr
	}

	body, appErr := renderTemplateFile(NotificationFile(name, NotificationBodyFile), name.String(), "notification", NotificationBodyFile, data)
	if appErr != nil {
		return nil, appErr
	}

	actionURL, appErr := renderTemplateFile(NotificationFile(name, NotificationActionURLFile), name.String(), "notification", NotificationActionURLFile, data)
	if appErr != nil {
		return nil, appErr
	}

	return &RenderedNotification{
		Title:     title,
		Body:      body,
		ActionURL: actionURL,
	}, nil
}

func renderTemplateFile(templatePath, name, kind, file string, data map[string]string) (string, *apperror.AppError) {
	tmpl, err := texttemplate.New(path.Base(file)).Option("missingkey=error").ParseFS(FS, templatePath)
	if err != nil {
		return "", apperror.New(apperror.CodeInternal, "failed to parse "+kind+" template").
			WithDetail("template", name).
			WithDetail("file", file).
			WithDetail("error", err.Error())
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", apperror.New(apperror.CodeValidation, "failed to render "+kind+" template").
			WithDetail("template", name).
			WithDetail("file", file).
			WithDetail("error", err.Error())
	}

	return out.String(), nil
}
