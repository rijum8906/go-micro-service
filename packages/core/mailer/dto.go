package mailer

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rijum8906/relay/packages/core/apperror"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("mail_address", validateMailAddress)
	validate.RegisterStructValidation(validateConfig, Config{})
	validate.RegisterStructValidation(validateEnvelope, Envelope{})
	validate.RegisterStructValidation(validateContent, Content{})
}

type Config struct {
	Host        string `validate:"required"`
	Port        int    `validate:"required,gt=0"`
	Username    string
	Password    string
	FromEmail   string `validate:"omitempty,email"`
	FromName    string
	UseTLS      bool
	UseStartTLS bool
	Timeout     time.Duration `validate:"gte=0"`
	Retries     int           `validate:"gte=0"`
}

func (c Config) Validate() *apperror.AppError {
	if err := validate.Struct(c); err != nil {
		return validationError("invalid mailer config", err)
	}

	return nil
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string `validate:"required"`
	ContentType string
	Data        []byte `validate:"required,min=1"`
}

func (a Attachment) Validate() *apperror.AppError {
	if err := validate.Struct(a); err != nil {
		return validationError("invalid mail attachment", err)
	}

	return nil
}

// Envelope contains the addresses required for an email
type Envelope struct {
	From mail.Address   `validate:"mail_address"`
	To   []mail.Address `validate:"dive,mail_address"`
	CC   []mail.Address `validate:"dive,mail_address"`
	BCC  []mail.Address `validate:"dive,mail_address"`
}

func (m Envelope) Validate() *apperror.AppError {
	if err := validate.Struct(m); err != nil {
		return validationError("invalid mail envelope", err)
	}

	return nil
}

func (m Envelope) Recipients() []mail.Address {
	recipients := make([]mail.Address, 0, len(m.To)+len(m.CC)+len(m.BCC))
	recipients = append(recipients, m.To...)
	recipients = append(recipients, m.CC...)
	recipients = append(recipients, m.BCC...)

	return recipients
}

// Content contains the content of an email
type Content struct {
	Subject     string `validate:"required"`
	HTML        string
	Text        string
	Priority    EmailPriority
	Attachments []Attachment `validate:"dive"`
	Headers     map[string]string
}

func (b Content) Validate() *apperror.AppError {
	if err := validate.Struct(b); err != nil {
		return validationError("invalid mail content", err)
	}

	return nil
}

func (b Content) HasContent() bool {
	return strings.TrimSpace(b.HTML) != "" || strings.TrimSpace(b.Text) != ""
}

type Message struct {
	Envelope Envelope `validate:"required"`
	Content  Content  `validate:"required"`
}

func (m Message) Validate() *apperror.AppError {
	if err := validate.Struct(m); err != nil {
		return validationError("invalid mail message", err)
	}

	return nil
}

func validateMailAddress(fl validator.FieldLevel) bool {
	addr, ok := fl.Field().Interface().(mail.Address)
	if !ok {
		return false
	}
	if strings.TrimSpace(addr.Address) == "" {
		return false
	}

	_, err := mail.ParseAddress(addr.String())
	return err == nil
}

func validateEnvelope(sl validator.StructLevel) {
	envelope, ok := sl.Current().Interface().(Envelope)
	if !ok {
		return
	}
	if len(envelope.Recipients()) == 0 {
		sl.ReportError(envelope.To, "To", "to", "recipients_required", "")
	}
}

func validateConfig(sl validator.StructLevel) {
	config, ok := sl.Current().Interface().(Config)
	if !ok {
		return
	}
	if (config.Username == "") != (config.Password == "") {
		sl.ReportError(config.Username, "Username", "username", "auth_pair_required", "")
	}
	if config.UseTLS && config.UseStartTLS {
		sl.ReportError(config.UseTLS, "UseTLS", "useTLS", "tls_mode_conflict", "")
	}
}

func validateContent(sl validator.StructLevel) {
	content, ok := sl.Current().Interface().(Content)
	if !ok {
		return
	}
	if !content.HasContent() {
		sl.ReportError(content.HTML, "HTML", "html", "content_required", "")
	}
}

func validationError(message string, err error) *apperror.AppError {
	appErr := apperror.New(apperror.CodeValidation, message)

	var validationErrs validator.ValidationErrors
	if ok := errors.As(err, &validationErrs); ok {
		for _, validationErr := range validationErrs {
			appErr.WithDetail(validationErr.Field(), validationErr.Tag())
		}
		return appErr
	}

	return appErr.WithDetail("error", err.Error())
}
