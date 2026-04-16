// Package broker
package broker

import (
	"fmt"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/nats"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/internal/constants"
	"github.com/rijum8906/relay/services/notification-service/internal/services/subscriber"
)

type SubscribeHandler struct {
	SubscriberService subscriber.Service
	NatsClient        *nats.Client
	templateManager   template.TemplateManager
	mailerCfg         *mailer.Config
}

func New(subService subscriber.Service, client *nats.Client, mailerCfg *mailer.Config) (*SubscribeHandler, *apperror.AppError) {
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

	return &SubscribeHandler{
		SubscriberService: subService,
		NatsClient:        client,
		mailerCfg:         mailerCfg,
		templateManager:   tm,
	}, nil
}

func (h *SubscribeHandler) Subscribe() *apperror.AppError {
	if h == nil {
		return apperror.ErrInternal.WithMessage("subscribe handler is not initialized")
	}

	if h.NatsClient == nil {
		return apperror.ErrInternal.WithMessage("nats client is not initialized")
	}

	if h.mailerCfg == nil {
		return apperror.ErrInternal.WithMessage("mailer config is not initialized")
	}

	if h.templateManager == nil {
		return apperror.ErrInternal.WithMessage("template manager is not initialized")
	}

	h.SubscriberService.SubscribeEmailVerificationJob(constants.DurableEmailVerification, constants.StreamEmailVerification)

	return nil
}
