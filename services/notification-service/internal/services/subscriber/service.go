package subscriber

import (
	"encoding/json"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/internal/constants"
	"github.com/rijum8906/relay/services/notification-service/internal/utils"
	"go.uber.org/zap"
)

func (s *service) SubscribeEmailVerificationJob(durable constants.DurableName, stream constants.StreamName) *apperror.AppError {
	sub, appErr := s.NATSClient.PullSubscribe(dto.JobEmailVerification, string(durable), string(stream))
	if appErr != nil {
		return appErr
	}

	for {
		msgs, err := sub.Fetch(5)
		if err != nil {
			continue // Timeout? Just try fetching again.
		}

		for _, msg := range msgs {
			var data dto.EmailVerificationDTO

			err := json.Unmarshal(msg.Data, &data)
			if err != nil {
				s.AppLogger.Error("error unmarshalling job", zap.Error(err), zap.String("details", appErr.Details[0].Message))
				continue
			}

			emailTemplate, appErr := s.TM.RenderToString(template.TemplateTypeEmailVerification, data)
			if appErr != nil {
				s.AppLogger.Error("error rendering email template", zap.Error(err), zap.String("details", appErr.Details[0].Message))
			}

			envelop, appErr := utils.ParseMailEnvelop(s.MailerCfg, data.ClientEmail)
			if appErr != nil {
				s.AppLogger.Error("error parsing mail envelop", zap.Error(err), zap.String("details", appErr.Details[0].Message))
			}

			appErr = mailer.Send(s.Mailer, mailer.Message{
				Envelope: envelop,
				Content: mailer.Content{
					HTML:        emailTemplate,
					Subject:     "Email Verification",
					Priority:    mailer.EmailPriorityHigh,
					Attachments: []mailer.Attachment{},
					Headers:     map[string]string{},
				},
			})
			if appErr != nil {
				s.AppLogger.Error("error sending email", zap.Error(err), zap.String("details", appErr.Details[0].Message))
			}

			msg.Ack()

		}
	}
}
