// Package subscriber
package subscriber

import (
	"fmt"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/nats"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/internal/constants"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

type Service interface {
	SubscribeEmailVerificationJob(durable constants.DurableName, stream constants.StreamName) *apperror.AppError
}

type service struct {
	NATSClient *nats.Client
	AppLogger  *zap.Logger
	TM         template.TemplateManager
	Mailer     *mail.Client
	MailerCfg  *mailer.Config
}

func New(client *nats.Client, stream string, logger *zap.Logger, mailerCfg mailer.Config) (Service, *apperror.AppError) {
	mailer, appErr := mailer.Connect(mailerCfg)
	if appErr != nil {
		return nil, appErr
	}

	tm, err := template.NewTemplateManagerWithCompanyInfo("packages/templates", &dto.CompanyInfo{
		Name:       "Relay",
		Emails:     []string{"UfNwO@example.com"},
		Addresses:  []string{"123 Main St, Anytown, USA"},
		WebsiteURL: "https://relay.com",
		SocialLinks: []dto.SocialLink{
			{
				Label: "Twitter",
				URL:   "https://twitter.com/relay",
			},
		},
	})
	if err != nil {
		fmt.Println("Error creating template manager:", err)
		return nil, apperror.ErrInternal.WithMessage("failed to create template manager").WithDetail("error", err.Error())
	}

	return &service{
		NATSClient: client,
		AppLogger:  logger,
		Mailer:     mailer,
		MailerCfg:  &mailerCfg,
		TM:         tm,
	}, nil
}
