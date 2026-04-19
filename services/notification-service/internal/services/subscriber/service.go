package subscriber

import (
"encoding/json"

"github.com/rijum8906/relay/packages/core/apperror"
"github.com/rijum8906/relay/packages/core/dto"
"github.com/rijum8906/relay/packages/core/mailer"
"github.com/rijum8906/relay/packages/core/template"
"github.com/rijum8906/relay/services/notification-service/internal/utils"
"go.uber.org/zap"
)

func (s *service) SubscribeEmailVerificationJob(consumerName string) *apperror.AppError {
sub, appErr := s.BrokerConsumerManager.PullSubscribe(dto.JobEmailVerification, consumerName)
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

if err := json.Unmarshal(msg.Data, &data); err != nil {
s.AppLogger.Error("error unmarshalling job", zap.Error(err))
// Nack so it can be retried if desired
if err2 := msg.Nak(); err2 != nil {
s.AppLogger.Error("failed to nak message", zap.Error(err2))
}
continue
}

emailTemplate, appErr := s.TM.RenderToString(template.TemplateTypeEmailVerification, data)
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

if appErr = mailer.SendWithConfig(*s.MailerCfg, mailer.Message{
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
_ = msg.Nak()
continue
}

if err := msg.Ack(); err != nil {
s.AppLogger.Error("failed to ack message", zap.Error(err))
}
}
}
}

func (s *service) SubscribeJobPasswordReset(consumerName string) *apperror.AppError {
sub, appErr := s.BrokerConsumerManager.PullSubscribe(dto.JobEmailPasswordReset, consumerName)
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

if appErr = mailer.SendWithConfig(*s.MailerCfg, mailer.Message{
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
