// Package email is email service
package email

import "github.com/rijum8906/relay/packages/core/template"

type Service interface {
	SendEmail(to string, subject string, template template.TemplateType)
}

type service struct {
	// TODO: add email service
}

func New() Service {
	return &service{}
}
