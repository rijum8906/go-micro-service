package dto

import (
	"net/mail"
	"net/url"
	"strconv"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/templates"
)

type EmailMessage struct {
	Meta        EmailMetadata           `json:"meta"`
	BodyContent map[string]string       `json:"body_content"`
	Attachments []*EmailAttachment      `json:"attachments"`
	Template    templates.EmailTemplate `json:"template"`
}

type EmailMetadata struct {
	JobSubject JobSubject       `json:"job_subject"`
	Sender     EmailSender      `json:"sender"`
	Recipients []EmailRecipient `json:"recipients"`
}

type EmailRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type EmailSender struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type EmailAttachment struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}

type NotificationChannel string

const (
	NotificationChannelPush    NotificationChannel = "push"
	NotificationChannelSMS     NotificationChannel = "sms"
	NotificationChannelWebhook NotificationChannel = "webhook"
)

type NotificationMessage struct {
	JobSubject  JobSubject               `json:"job_subject"`
	Channel     NotificationChannel      `json:"channel"`
	Recipient   NotificationRecipient    `json:"recipient"`
	Content     NotificationContent      `json:"content"`
	PushData    *PushNotificationData    `json:"push_data,omitempty"`
	SMSData     *SMSNotificationData     `json:"sms_data,omitempty"`
	WebhookData *WebhookNotificationData `json:"webhook_data,omitempty"`
}

type NotificationRecipient struct {
	UserID      string `json:"user_id,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	DeviceToken string `json:"device_token,omitempty"`
	WebhookURL  string `json:"webhook_url,omitempty"`
}

type NotificationContent struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	ActionURL string `json:"action_url,omitempty"`
}

type PushNotificationData struct {
	DeviceToken string `json:"device_token"`
	Title       string `json:"title"`
	Body        string `json:"body"`
}

type SMSNotificationData struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
}

type WebhookNotificationData struct {
	URL     string         `json:"url"`
	Event   string         `json:"event"`
	Payload map[string]any `json:"payload,omitempty"`
}

func (m EmailMessage) Validate() *apperror.AppError {
	if appErr := m.Meta.Validate(); appErr != nil {
		return appErr
	}

	if appErr := m.Template.Validate(); appErr != nil {
		return appErr
	}

	if len(m.BodyContent) == 0 {
		return validationError("body_content", "body content is required")
	}

	for i, attachment := range m.Attachments {
		if attachment == nil {
			return validationError("attachments", "attachment is required").WithDetail("attachment_index", strconv.Itoa(i))
		}

		if appErr := attachment.Validate(); appErr != nil {
			return appErr.WithDetail("attachment_index", strconv.Itoa(i))
		}
	}

	return nil
}

func (m EmailMetadata) Validate() *apperror.AppError {
	if !IsValidJobSubject(m.JobSubject.String()) || m.JobSubject.Domain() != string(DomainEmail) {
		return validationError("job_subject", "invalid email job subject")
	}

	if appErr := m.Sender.Validate(); appErr != nil {
		return appErr
	}

	if len(m.Recipients) == 0 {
		return validationError("recipients", "at least one recipient is required")
	}

	for i, recipient := range m.Recipients {
		if appErr := recipient.Validate(); appErr != nil {
			return appErr.WithDetail("recipient_index", strconv.Itoa(i))
		}
	}

	return nil
}

func (r EmailRecipient) Validate() *apperror.AppError {
	return validateEmailField("recipient.email", r.Email)
}

func (s EmailSender) Validate() *apperror.AppError {
	return validateEmailField("sender.email", s.Email)
}

func (a EmailAttachment) Validate() *apperror.AppError {
	if a.FileName == "" {
		return validationError("attachment.file_name", "file name is required")
	}
	if a.ContentType == "" {
		return validationError("attachment.content_type", "content type is required")
	}
	if appErr := validateURLField("attachment.url", a.URL); appErr != nil {
		return appErr
	}

	return nil
}

func (m NotificationMessage) Validate() *apperror.AppError {
	if !IsValidJobSubject(m.JobSubject.String()) || m.JobSubject.Domain() != string(DomainNotification) {
		return validationError("job_subject", "invalid notification job subject")
	}

	if appErr := m.Channel.Validate(); appErr != nil {
		return appErr
	}

	if appErr := m.Recipient.Validate(m.Channel); appErr != nil {
		return appErr
	}

	if appErr := m.Content.Validate(); appErr != nil {
		return appErr
	}

	switch m.Channel {
	case NotificationChannelPush:
		if m.JobSubject != JobNotificationPush {
			return validationError("job_subject", "job subject does not match push notification channel")
		}
		if m.PushData == nil {
			return validationError("push_data", "push data is required")
		}
		return m.PushData.Validate(m.Recipient)
	case NotificationChannelSMS:
		if m.JobSubject != JobNotificationSMS {
			return validationError("job_subject", "job subject does not match sms notification channel")
		}
		if m.SMSData == nil {
			return validationError("sms_data", "sms data is required")
		}
		return m.SMSData.Validate(m.Recipient)
	case NotificationChannelWebhook:
		if m.JobSubject != JobNotificationWebhook {
			return validationError("job_subject", "job subject does not match webhook notification channel")
		}
		if m.WebhookData == nil {
			return validationError("webhook_data", "webhook data is required")
		}
		return m.WebhookData.Validate(m.Recipient)
	default:
		return validationError("channel", "unsupported notification channel")
	}
}

func (c NotificationChannel) Validate() *apperror.AppError {
	switch c {
	case NotificationChannelPush, NotificationChannelSMS, NotificationChannelWebhook:
		return nil
	default:
		return validationError("channel", "unsupported notification channel")
	}
}

func (r NotificationRecipient) Validate(channel NotificationChannel) *apperror.AppError {
	switch channel {
	case NotificationChannelPush:
		if r.DeviceToken == "" {
			return validationError("recipient.device_token", "device token is required")
		}
	case NotificationChannelSMS:
		if r.PhoneNumber == "" {
			return validationError("recipient.phone_number", "phone number is required")
		}
	case NotificationChannelWebhook:
		if appErr := validateURLField("recipient.webhook_url", r.WebhookURL); appErr != nil {
			return appErr
		}
	}

	return nil
}

func (c NotificationContent) Validate() *apperror.AppError {
	if c.Body == "" {
		return validationError("content.body", "body is required")
	}

	if c.ActionURL != "" {
		if appErr := validateURLField("content.action_url", c.ActionURL); appErr != nil {
			return appErr
		}
	}

	return nil
}

func (d PushNotificationData) Validate(recipient NotificationRecipient) *apperror.AppError {
	if d.DeviceToken == "" {
		return validationError("push_data.device_token", "device token is required")
	}
	if d.Title == "" {
		return validationError("push_data.title", "title is required")
	}
	if d.Body == "" {
		return validationError("push_data.body", "body is required")
	}
	if recipient.DeviceToken != "" && recipient.DeviceToken != d.DeviceToken {
		return validationError("push_data.device_token", "push device token does not match recipient device token")
	}

	return nil
}

func (d SMSNotificationData) Validate(recipient NotificationRecipient) *apperror.AppError {
	if d.PhoneNumber == "" {
		return validationError("sms_data.phone_number", "phone number is required")
	}
	if d.Message == "" {
		return validationError("sms_data.message", "message is required")
	}
	if recipient.PhoneNumber != "" && recipient.PhoneNumber != d.PhoneNumber {
		return validationError("sms_data.phone_number", "sms phone number does not match recipient phone number")
	}

	return nil
}

func (d WebhookNotificationData) Validate(recipient NotificationRecipient) *apperror.AppError {
	if appErr := validateURLField("webhook_data.url", d.URL); appErr != nil {
		return appErr
	}
	if d.Event == "" {
		return validationError("webhook_data.event", "event is required")
	}
	if recipient.WebhookURL != "" && recipient.WebhookURL != d.URL {
		return validationError("webhook_data.url", "webhook url does not match recipient webhook url")
	}

	return nil
}

func validateEmailField(field, value string) *apperror.AppError {
	if value == "" {
		return validationError(field, "email is required")
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return validationError(field, "invalid email address")
	}

	return nil
}

func validateURLField(field, value string) *apperror.AppError {
	if value == "" {
		return validationError(field, "url is required")
	}

	parsedURL, err := url.ParseRequestURI(value)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return validationError(field, "invalid url")
	}

	return nil
}

func validationError(field, message string) *apperror.AppError {
	return apperror.New(apperror.CodeValidation, message).WithDetail(field, message)
}
