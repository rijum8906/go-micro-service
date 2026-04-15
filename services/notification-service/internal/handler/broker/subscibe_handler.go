// Package broker
package broker

import (
	"encoding/json"
	"fmt"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/nats"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/internal/services/email"
	"github.com/rijum8906/relay/services/notification-service/internal/utils"
	"github.com/wneessen/go-mail"
)

type SubscribeHandler struct {
	EmailService    email.Service
	NatsClient      *nats.Client
	templateManager template.TemplateManager
	mailerCfg       *mailer.Config
	mailerClient    *mail.Client
}

func New(emailService email.Service, client *nats.Client, mailerCfg *mailer.Config) (*SubscribeHandler, *apperror.AppError) {
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
		EmailService:    emailService,
		NatsClient:      client,
		mailerCfg:       mailerCfg,
		templateManager: tm,
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

	client, appErr := mailer.Connect(*h.mailerCfg)
	if appErr != nil {
		return appErr
	}
	h.mailerClient = client

	_, appErr = h.NatsClient.Subscribe(dto.JobEmailVerification, h.handlerEmailVerification)
	if appErr != nil {
		return appErr
	}
	_, appErr = h.NatsClient.Subscribe(dto.JobEmailPasswordReset, h.handlerPasswordReset)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (h *SubscribeHandler) handlerEmailVerification(raw []byte) {
	var data dto.EmailVerificationDTO
	err := json.Unmarshal(raw, &data)
	if err != nil {
		// TODO: save to some logs
		fmt.Println("Error unmarshalling job:", err)
		return
	}

	emailTemplate, appErr := h.templateManager.RenderToString(template.TemplateTypeEmailVerification, data)
	if appErr != nil {
		// TODO: save to some logs
		fmt.Println("Error rendering email template:", appErr.Details)
		return
	}

	envelop, appErr := utils.ParseMailEnvelop(h.mailerCfg, data.ClientEmail)
	if appErr != nil {
		// TODO: save to some logs
		fmt.Println("Error parsing mail envelop:", appErr)
		return
	}

	appErr = mailer.Send(h.mailerClient, mailer.Message{
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
		// TODO: save to some logs
		fmt.Println("Error sending email:", appErr.Details)
		return
	}
}

func (h *SubscribeHandler) handlerPasswordReset(raw []byte) {
	var data dto.PasswordResetDTO
	err := json.Unmarshal(raw, &data)
	if err != nil {
		// TODO: save to some logs
		fmt.Println("Error unmarshalling job:", err)
		return
	}

	emailTemplate, appErr := h.templateManager.RenderToString(template.TemplateTypeEmailPasswordReset, data)
	if appErr != nil {
		// TODO: save to some logs
		fmt.Println("Error rendering email template:", appErr.Details)
		return
	}

	envelop, appErr := utils.ParseMailEnvelop(h.mailerCfg, data.ClientEmail)
	if appErr != nil {
		// TODO: save to some logs
		fmt.Println("Error parsing mail envelop:", appErr)
		return
	}

	appErr = mailer.Send(h.mailerClient, mailer.Message{
		Envelope: envelop,
		Content: mailer.Content{
			HTML:        emailTemplate,
			Subject:     "Password Reset",
			Priority:    mailer.EmailPriorityHigh,
			Attachments: []mailer.Attachment{},
			Headers:     map[string]string{},
		},
	})
	if appErr != nil {
		// TODO: save to some logs
		fmt.Println("Error sending email:", appErr.Details)
		return
	}
}
