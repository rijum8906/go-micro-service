package subscriber

import (
	"encoding/json"
	"fmt"
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

// SubscribeUserAuthEmailJobs - ONE consumer for ALL email job types
func (s *service) SubscribeUserAuthEmailJobs(consumerName string) *apperror.AppError {
	sub, appErr := s.BrokerConsumerManager.PullSubscribe(jobs.GetSubdomainWildcard(jobs.JobUserRequestedEmailVerification), consumerName)
	if appErr != nil {
		fmt.Println(appErr.Details)
	}

	batchSize := 5
	waitTime := 30 * time.Second

	for {
		msgs, err := sub.Fetch(batchSize, nats.MaxWait(waitTime))
		if err != nil {
			if err == nats.ErrTimeout {
				time.Sleep(1 * time.Second)
				continue
			}
			s.AppLogger.Error("failed to fetch messages", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		for _, msg := range msgs {
			s.routeAndProcess(msg)
		}
	}
}

func (s *service) routeAndProcess(msg *nats.Msg) {
	// Check retry count
	metadata, err := msg.Metadata()
	if err == nil && metadata.NumDelivered > 3 {
		s.AppLogger.Error("max retries exceeded, discarding",
			zap.String("subject", msg.Subject),
			zap.Int("retries", int(metadata.NumDelivered)))
		_ = msg.Term()
		return
	}

	switch msg.Subject {
	case jobs.JobUserRequestedPasswordReset:
		s.processPasswordReset(msg)
	case jobs.JobUserRequestedEmailVerification:
		s.processEmailVerification(msg)
	default:
		s.AppLogger.Warn("unknown job subject", zap.String("subject", msg.Subject))
		_ = msg.Ack() // Ack unknown to avoid infinite retries
	}
}

func (s *service) processEmailVerification(msg *nats.Msg) {
	var data dto.EmailVerificationDTO

	if err := json.Unmarshal(msg.Data, &data); err != nil {
		s.AppLogger.Error("error unmarshalling email verification",
			zap.Error(err),
			zap.String("data", string(msg.Data)))
		_ = msg.NakWithDelay(5 * time.Second)
		return
	}

	// Render template
	emailTemplate, err := s.TM.RenderToString(template.TemplateTypeEmailVerification, data)
	if err != nil {
		s.AppLogger.Error("error rendering email verification template",
			zap.Error(err),
			zap.String("details", err.Error()))
		_ = msg.NakWithDelay(10 * time.Second)
		return
	}

	// Process email
	s.sendEmail(msg, emailTemplate, "Email Verification", data.BaseEmailDTO)
}

func (s *service) processPasswordReset(msg *nats.Msg) {
	var data dto.PasswordResetDTO

	if err := json.Unmarshal(msg.Data, &data); err != nil {
		s.AppLogger.Error("error unmarshalling password reset",
			zap.Error(err),
			zap.String("data", string(msg.Data)))
		_ = msg.NakWithDelay(5 * time.Second)
		return
	}

	emailTemplate, err := s.TM.RenderToString(template.TemplateTypeEmailPasswordReset, data)
	if err != nil {
		s.AppLogger.Error("error rendering password reset template",
			zap.Error(err),
			zap.String("details", err.Error()))
		_ = msg.NakWithDelay(10 * time.Second)
		return
	}

	s.sendEmail(msg, emailTemplate, "Password Reset", data.BaseEmailDTO)
}

func (s *service) sendEmail(msg *nats.Msg, emailTemplate string, subject string, baseData dto.BaseEmailDTO) {
	envelop := utils.ParseMailEnvelop(s.MailerCfg, baseData.ClientEmail)

	// Send email
	if err := mailer.SendWithConfig(*s.MailerCfg, mailer.Message{
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
		s.AppLogger.Error("error sending email",
			zap.Error(err),
			zap.String("email", baseData.ClientEmail))
		_ = msg.NakWithDelay(30 * time.Second)
		return
	}

	// Success
	if err := msg.Ack(); err != nil {
		s.AppLogger.Error("failed to ack message", zap.Error(err))
		return
	}

	s.AppLogger.Info("email sent successfully",
		zap.String("subject", subject),
		zap.String("email", baseData.ClientEmail))
}
