// Package subscriber
package subscriber

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

type Service interface {
	SubscribeUserAuthEmailJobs(consumerName string) *apperror.AppError
}

type service struct {
	BrokerConsumerManager broker.Subscriber
	AppLogger             *zap.Logger
	TM                    template.TemplateManager
	Mailer                *mail.Client
	MailerCfg             *mailer.Config
}

func New(client broker.Subscriber, logger *zap.Logger, mailerCfg mailer.Config, tm template.TemplateManager) (Service, *apperror.AppError) {
	mailer, appErr := mailer.Connect(mailerCfg)
	if appErr != nil {
		return nil, appErr
	}

	return &service{
		BrokerConsumerManager: client,
		AppLogger:             logger,
		Mailer:                mailer,
		MailerCfg:             &mailerCfg,
		TM:                    tm,
	}, nil
}
