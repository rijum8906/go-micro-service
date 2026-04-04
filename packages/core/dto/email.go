package dto

import (
	"net/mail"
	"net/url"
	"strconv"

	"github.com/rijum8906/relay/packages/core/apperror"
)

type EmailMessage struct {
	JobSubject        JobSubject              `json:"job_subject"`
	Sender            EmailSender             `json:"sender"`
	Recipients        []EmailRecipient        `json:"recipients"`
	Content           EmailContent            `json:"content"`
	Attachments       []EmailAttachment       `json:"attachments,omitempty"`
	VerificationData  *VerificationEmailData  `json:"verification_data,omitempty"`
	PasswordResetData *PasswordResetEmailData `json:"password_reset_data,omitempty"`
	WelcomeData       *WelcomeEmailData       `json:"welcome_data,omitempty"`
}

type EmailRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type EmailSender struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type EmailContent struct {
	SubjectLine  string `json:"subject_line"`
	TemplateName string `json:"template_name,omitempty"`
	TextBody     string `json:"text_body,omitempty"`
	HTMLBody     string `json:"html_body,omitempty"`
}

type EmailAttachment struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}

type VerificationEmailData struct {
	UserID            string `json:"user_id"`
	Email             string `json:"email"`
	FirstName         string `json:"first_name,omitempty"`
	VerificationToken string `json:"verification_token"`
	VerificationURL   string `json:"verification_url"`
}

type PasswordResetEmailData struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	FirstName  string `json:"first_name,omitempty"`
	ResetToken string `json:"reset_token"`
	ResetURL   string `json:"reset_url"`
}

type WelcomeEmailData struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LoginURL  string `json:"login_url"`
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

	if appErr := m.Content.Validate(); appErr != nil {
		return appErr
	}

	for i, attachment := range m.Attachments {
		if appErr := attachment.Validate(); appErr != nil {
			return appErr.WithDetail("attachment_index", strconv.Itoa(i))
		}
	}

	switch m.JobSubject {
	case JobEmailVerification:
		if m.VerificationData == nil {
			return validationError("verification_data", "verification data is required")
		}
		return m.VerificationData.Validate()
	case JobEmailPasswordReset:
		if m.PasswordResetData == nil {
			return validationError("password_reset_data", "password reset data is required")
		}
		return m.PasswordResetData.Validate()
	case JobEmailWelcome:
		if m.WelcomeData == nil {
			return validationError("welcome_data", "welcome data is required")
		}
		return m.WelcomeData.Validate()
	default:
		return nil
	}
}

func (r EmailRecipient) Validate() *apperror.AppError {
	return validateEmailField("recipient.email", r.Email)
}

func (s EmailSender) Validate() *apperror.AppError {
	return validateEmailField("sender.email", s.Email)
}

func (c EmailContent) Validate() *apperror.AppError {
	if c.SubjectLine == "" {
		return validationError("content.subject_line", "subject line is required")
	}

	if c.TemplateName == "" && c.TextBody == "" && c.HTMLBody == "" {
		return validationError("content", "template name, text body, or html body is required")
	}

	return nil
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

func (d VerificationEmailData) Validate() *apperror.AppError {
	if d.UserID == "" {
		return validationError("verification_data.user_id", "user id is required")
	}
	if appErr := validateEmailField("verification_data.email", d.Email); appErr != nil {
		return appErr
	}
	if d.VerificationToken == "" {
		return validationError("verification_data.verification_token", "verification token is required")
	}
	if appErr := validateURLField("verification_data.verification_url", d.VerificationURL); appErr != nil {
		return appErr
	}

	return nil
}

func (d PasswordResetEmailData) Validate() *apperror.AppError {
	if d.UserID == "" {
		return validationError("password_reset_data.user_id", "user id is required")
	}
	if appErr := validateEmailField("password_reset_data.email", d.Email); appErr != nil {
		return appErr
	}
	if d.ResetToken == "" {
		return validationError("password_reset_data.reset_token", "reset token is required")
	}
	if appErr := validateURLField("password_reset_data.reset_url", d.ResetURL); appErr != nil {
		return appErr
	}

	return nil
}

func (d WelcomeEmailData) Validate() *apperror.AppError {
	if d.UserID == "" {
		return validationError("welcome_data.user_id", "user id is required")
	}
	if appErr := validateEmailField("welcome_data.email", d.Email); appErr != nil {
		return appErr
	}
	if appErr := validateURLField("welcome_data.login_url", d.LoginURL); appErr != nil {
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
	return apperror.New(apperror.TypeValidation, apperror.CodeValidation, message).WithDetail(field, message)
}
