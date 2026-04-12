package mailer

import (
	"net/mail"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("mail_address", validateMailAddress)
	validate.RegisterStructValidation(validateEnvelope, Envelope{})
	validate.RegisterStructValidation(validateContent, Content{})
}

type Config struct {
	Host        string `validate:"required"`
	Port        int    `validate:"required,gt=0"`
	Username    string `validate:"required"`
	Password    string `validate:"required"`
	FromEmail   string `validate:"required,email"`
	FromName    string
	UseTLS      bool
	UseStartTLS bool
	Timeout     time.Duration `validate:"gte=0"`
	Retries     int           `validate:"gte=0"`
}

func (c Config) Validate() error {
	return validate.Struct(c)
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string `validate:"required"`
	ContentType string
	Data        []byte `validate:"required,min=1"`
}

func (a Attachment) Validate() error {
	return validate.Struct(a)
}

// Envelope contains the addresses required for an email
type Envelope struct {
	From mail.Address   `validate:"mail_address"`
	To   []mail.Address `validate:"dive,mail_address"`
	CC   []mail.Address `validate:"dive,mail_address"`
	BCC  []mail.Address `validate:"dive,mail_address"`
}

func (m Envelope) Validate() error {
	return validate.Struct(m)
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

func (b Content) Validate() error {
	return validate.Struct(b)
}

func (b Content) HasContent() bool {
	return strings.TrimSpace(b.HTML) != "" || strings.TrimSpace(b.Text) != ""
}

type Message struct {
	Envelope Envelope `validate:"required"`
	Content  Content  `validate:"required"`
}

func (m Message) Validate() error {
	return validate.Struct(m)
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

func validateContent(sl validator.StructLevel) {
	content, ok := sl.Current().Interface().(Content)
	if !ok {
		return
	}
	if !content.HasContent() {
		sl.ReportError(content.HTML, "HTML", "html", "content_required", "")
	}
}
