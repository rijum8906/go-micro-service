package subscriber

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/jobs"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/internal/utils"
	"go.uber.org/zap"
)

func (s *service) SubscribeEmailVerificationJob(consumerName string) *apperror.AppError {
	sub, appErr := s.BrokerConsumerManager.PullSubscribe(jobs.JobUserRequestedEmailVerification, consumerName)
	if appErr != nil {
		return appErr
	}

	// Set max messages and timeout
	batchSize := 5
	waitTime := 30 * time.Second

	for {
		// Add MaxWait to prevent indefinite block
		msgs, err := sub.Fetch(batchSize, nats.MaxWait(waitTime))
		if err != nil {
			if err == nats.ErrTimeout {
				// No messages in queue, that's normal
				time.Sleep(1 * time.Second) // Don't hammer the server
				continue
			}
			s.AppLogger.Error("failed to fetch messages", zap.Error(err))
			time.Sleep(5 * time.Second) // Backoff on error
			continue
		}

		for _, msg := range msgs {
			s.processEmailVerification(msg)
		}
	}
}

func (s *service) processEmailVerification(msg *nats.Msg) {
	var data dto.EmailVerificationDTO

	if err := json.Unmarshal(msg.Data, &data); err != nil {
		s.AppLogger.Error("error unmarshalling job", zap.Error(err))
		// Nak with delay (retry after 5 seconds)
		_ = msg.NakWithDelay(5 * time.Second)
		return
	}

	emailTemplate, appErr := s.TM.RenderToString(template.TemplateTypeEmailVerification, data)
	if appErr != nil {
		s.AppLogger.Error("error rendering email template", zap.String("details", appErr.Error()))
		_ = msg.NakWithDelay(10 * time.Second) // Give more time for template issue
		return
	}

	envelop, appErr := utils.ParseMailEnvelop(s.MailerCfg, data.ClientEmail)
	if appErr != nil {
		s.AppLogger.Error("error parsing mail envelop", zap.String("details", appErr.Error()))
		// This might be a permanent error (invalid email) - consider Term()
		if strings.Contains(appErr.Error(), "invalid email") {
			_ = msg.Term() // Permanently discard
			s.AppLogger.Warn("permanently failed message", zap.String("subject", msg.Subject))
		} else {
			_ = msg.NakWithDelay(10 * time.Second)
		}
		return
	}

	if appErr = mailer.Send(s.Mailer, mailer.Message{
		Envelope: envelop,
		Content: mailer.Content{
			HTML:        emailTemplate,
			Subject:     "Email Verification",
			Priority:    mailer.EmailPriorityHigh,
			Attachments: []mailer.Attachment{},
			Headers:     map[string]string{},
		},
	}); appErr != nil {
		s.AppLogger.Error("error sending email", zap.String("details", appErr.Error()))
		// Mailer error - retry with backoff
		_ = msg.NakWithDelay(30 * time.Second)
		return
	}

	// Success!
	if err := msg.Ack(); err != nil {
		s.AppLogger.Error("failed to ack message", zap.Error(err))
	}

	s.AppLogger.Info("email verification sent", zap.String("email", data.ClientEmail))
}

func (s *service) SubscribeJobPasswordReset(consumerName string) *apperror.AppError {
	sub, appErr := s.BrokerConsumerManager.PullSubscribe(jobs.JobUserRequestedPasswordReset, consumerName)
	if appErr != nil {
		return appErr
	}

	count := 0

	for {
		msgs, err := sub.Fetch(5)
		if err != nil {
			continue // Timeout? Just try fetching again.
		}

		for _, msg := range msgs {
			var data dto.PasswordResetDTO

			if err := json.Unmarshal(msg.Data, &data); err != nil {
				s.AppLogger.Error("error unmarshalling job", zap.Error(err))
				_ = msg.Nak()
				continue
			}

			emailTemplate, appErr := s.TM.RenderToString(template.TemplateTypeEmailPasswordReset, data)
			if appErr != nil {
				s.AppLogger.Error("error rendering email template", zap.String("details", appErr.Error()))
				_ = msg.Nak()
				continue
			}

			envelop, appErr := utils.ParseMailEnvelop(s.MailerCfg, data.ClientEmail)
			if appErr != nil {
				s.AppLogger.Error("error parsing mail envelop", zap.String("details", appErr.Error()))
				_ = msg.Nak()
				continue
			}

			if appErr = mailer.Send(s.Mailer, mailer.Message{
				Envelope: envelop,
				Content: mailer.Content{
					HTML:        emailTemplate,
					Subject:     "Password Reset",
					Priority:    mailer.EmailPriorityHigh,
					Attachments: []mailer.Attachment{},
					Headers:     map[string]string{},
				},
			}); appErr != nil {
				s.AppLogger.Error("error sending email", zap.String("details", appErr.Error()))
				_ = msg.Nak()
				continue
			}

			count++
			s.AppLogger.Info("sent password reset email", zap.Int("count", count))

			if err := msg.Ack(); err != nil {
				s.AppLogger.Error("failed to ack message", zap.Error(err))
			}
		}
	}
}
