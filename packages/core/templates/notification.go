package templates

import (
	"path"

	"github.com/rijum8906/relay/packages/core/apperror"
)

const (
	NotificationBaseDir       = "notification"
	NotificationTitleFile     = "title.tmpl"
	NotificationBodyFile      = "body.tmpl"
	NotificationActionURLFile = "action_url.tmpl"
)

type NotificationTemplate string

const (
	NotificationTemplatePush    NotificationTemplate = "push"
	NotificationTemplateSMS     NotificationTemplate = "sms"
	NotificationTemplateWebhook NotificationTemplate = "webhook"
)

var notificationTemplates = []NotificationTemplate{
	NotificationTemplatePush,
	NotificationTemplateSMS,
	NotificationTemplateWebhook,
}

var notificationTemplateSet = map[NotificationTemplate]struct{}{
	NotificationTemplatePush:    {},
	NotificationTemplateSMS:     {},
	NotificationTemplateWebhook: {},
}

func (t NotificationTemplate) String() string {
	return string(t)
}

func (t NotificationTemplate) Validate() *apperror.AppError {
	if t == "" {
		return validationError("template", "template is required")
	}

	if !NotificationExists(t) {
		return validationError("template", "invalid notification template")
	}

	return nil
}

func NotificationNames() []NotificationTemplate {
	return append([]NotificationTemplate(nil), notificationTemplates...)
}

func NotificationExists(name NotificationTemplate) bool {
	_, ok := notificationTemplateSet[name]
	return ok
}

func NotificationDir(name NotificationTemplate) string {
	return path.Join(NotificationBaseDir, name.String())
}

func NotificationFile(name NotificationTemplate, file string) string {
	return path.Join(NotificationDir(name), file)
}

func NotificationFiles(name NotificationTemplate) []string {
	return []string{
		NotificationFile(name, NotificationTitleFile),
		NotificationFile(name, NotificationBodyFile),
		NotificationFile(name, NotificationActionURLFile),
	}
}
