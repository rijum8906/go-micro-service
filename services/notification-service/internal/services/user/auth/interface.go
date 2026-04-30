package userauth

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/app"
	"go.uber.org/zap"
)

type UserAuthEmailService struct {
	// Core
	BrokerClient    broker.Client
	TemplateManager template.TemplateManager

	// Utilities
	Logger *zap.Logger

	// Config
	ConsumerInfo *nats.ConsumerInfo
	MailerConfig mailer.Config
}

func New() (*UserAuthEmailService, *apperror.AppError) {
	application, appErr := app.GetInstance()
	if appErr != nil {
		return nil, appErr
	}

	return &UserAuthEmailService{
		BrokerClient:    application.BrokerClient(),
		TemplateManager: application.TemplateManager(),
		Logger:          application.Logger(),
		MailerConfig:    application.MailerConfig(),
	}, nil
}
