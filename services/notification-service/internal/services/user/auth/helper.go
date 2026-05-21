package userauth

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/services/notification-service/internal/utils"
)

func (s *UserAuthEmailService) sendEmail(msg *nats.Msg, emailTemplate string, subject string, baseData dto.BaseEmailDTO) *apperror.AppError {
	envelop := utils.ParseMailEnvelop(&s.MailerConfig, baseData.ClientEmail)

	// Send email
	if err := mailer.SendWithConfig(s.MailerConfig, mailer.Message{
		Envelope: envelop,
		Content: mailer.Content{
			HTML:        emailTemplate,
			Subject:     subject,
			Priority:    mailer.EmailPriorityHigh,
			Attachments: []mailer.Attachment{},
			Headers: map[string]string{
				"X-Email-Type": subject,
			},
		},
	}); err != nil {
		return apperror.ErrThirdParty.WithMessage("error sending email").WithDetail("error", err.Error())
	}

	// Success
	if err := msg.Ack(); err != nil {
		return apperror.ErrInternal.WithMessage("failed to ack message").WithDetail("error", err.Error())
	}

	return nil
}
