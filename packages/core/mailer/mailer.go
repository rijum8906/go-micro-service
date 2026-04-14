// Package mailer
package mailer

import (
	"bytes"
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/rijum8906/relay/packages/core/apperror"
	gomail "github.com/wneessen/go-mail"
)

type EmailPriority int

const (
	EmailPriorityLow EmailPriority = iota
	EmailPriorityNormal
	EmailPriorityHigh
)

func Connect(config Config) (*gomail.Client, *apperror.AppError) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	client, err := gomail.NewClient(config.Host, buildClientOptions(config)...)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to create smtp client").
			WithDetail("error", err.Error())
	}

	ctx, cancel := connectContext(config.Timeout)
	defer cancel()

	if err := client.DialWithContext(ctx); err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to connect smtp client").
			WithDetail("error", err.Error())
	}

	return client, nil
}

func Send(client *gomail.Client, message Message) *apperror.AppError {
	if client == nil {
		return apperror.New(apperror.CodeInternal, "mail client is not initialized")
	}
	if err := message.Validate(); err != nil {
		return err
	}

	msg, err := buildMessage(message)
	if err != nil {
		return err
	}

	if err := client.Send(msg); err != nil {
		return apperror.New(apperror.CodeThirdParty, "failed to send email").
			WithDetail("error", err.Error())
	}

	return nil
}

func SendWithConfig(config Config, message Message) *apperror.AppError {
	client, err := Connect(config)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	return Send(client, message)
}

func buildClientOptions(config Config) []gomail.Option {
	opts := []gomail.Option{gomail.WithPort(config.Port)}

	switch {
	case config.UseTLS:
		opts = append(opts, gomail.WithSSL())
	case config.UseStartTLS:
		opts = append(opts, gomail.WithTLSPolicy(gomail.TLSMandatory))
	default:
		opts = append(opts, gomail.WithTLSPolicy(gomail.NoTLS))
	}

	if config.Username != "" {
		opts = append(opts,
			gomail.WithUsername(config.Username),
			gomail.WithPassword(config.Password),
			gomail.WithSMTPAuth(gomail.SMTPAuthAutoDiscover),
		)
	}

	return opts
}

func buildMessage(message Message) (*gomail.Msg, *apperror.AppError) {
	msg := gomail.NewMsg()
	if msg == nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to create email message")
	}

	msg.FromMailAddress(message.Envelope.From)
	msg.ToMailAddress(mailAddresses(message.Envelope.To)...)
	msg.CcMailAddress(mailAddresses(message.Envelope.CC)...)
	msg.BccMailAddress(mailAddresses(message.Envelope.BCC)...)
	msg.Subject(message.Content.Subject)
	msg.SetDate()
	msg.SetMessageID()
	msg.SetImportance(toImportance(message.Content.Priority))

	if err := setBodies(msg, message.Content); err != nil {
		return nil, err
	}

	for key, value := range message.Content.Headers {
		msg.SetGenHeader(gomail.Header(key), value)
	}

	for _, attachment := range message.Content.Attachments {
		opts := []gomail.FileOption{gomail.WithFileName(attachment.Filename)}
		if attachment.ContentType != "" {
			opts = append(opts, gomail.WithFileContentType(gomail.ContentType(attachment.ContentType)))
		}
		if err := msg.AttachReader(attachment.Filename, bytes.NewReader(attachment.Data), opts...); err != nil {
			return nil, apperror.New(apperror.CodeInternal, "failed to attach email file").
				WithDetail("filename", attachment.Filename).
				WithDetail("error", err.Error())
		}
	}

	return msg, nil
}

func setBodies(msg *gomail.Msg, content Content) *apperror.AppError {
	text := strings.TrimSpace(content.Text)
	html := strings.TrimSpace(content.HTML)

	switch {
	case text != "" && html != "":
		msg.SetBodyString(gomail.TypeTextPlain, content.Text)
		msg.AddAlternativeString(gomail.TypeTextHTML, content.HTML)
	case text != "":
		msg.SetBodyString(gomail.TypeTextPlain, content.Text)
	case html != "":
		msg.SetBodyString(gomail.TypeTextHTML, content.HTML)
	default:
		return apperror.New(apperror.CodeValidation, "message body must include html or text content")
	}

	return nil
}

func mailAddresses(addrs []*mail.Address) []*mail.Address {
	if len(addrs) == 0 {
		return nil
	}

	result := make([]*mail.Address, 0, len(addrs))
	for idx := range addrs {
		addr := addrs[idx]
		result = append(result, addr)
	}

	return result
}

func toImportance(priority EmailPriority) gomail.Importance {
	switch priority {
	case EmailPriorityLow:
		return gomail.ImportanceLow
	case EmailPriorityHigh:
		return gomail.ImportanceHigh
	default:
		return gomail.ImportanceNormal
	}
}

func connectContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(context.Background())
	}

	return context.WithTimeout(context.Background(), timeout)
}
