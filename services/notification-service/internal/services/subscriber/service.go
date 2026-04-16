package subscriber

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
)

type service struct {
	emailService    email.Service
	natsClient      *nats.Client
	templateManager template.TemplateManager
	mailerCfg       *mailer.Config
}

func New(emailService email.Service, natsClient *nats.Client, mailerCfg *mailer.Config) (Service, *apperror.AppError) {
	if emailService == nil {
		return nil, apperror.ErrInternal.WithMessage("email service is not initialized")
	}
	if natsClient == nil {
		return nil, apperror.ErrInternal.WithMessage("nats client  is not initialized")
	}
	if mailerCfg == nil {
		return nil, apperror.ErrInternal.WithMessage("mailer  is not initialized")
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
		emailService:    emailService,
		natsClient:      natsClient,
		templateManager: tm,
		mailerCfg:       mailerCfg,
	}, nil
}

func (h *service) Subscribe() *apperror.AppError {
	if h == nil {
		return apperror.ErrInternal.WithMessage("subscribe handler is not initialized")
	}

	if h.natsClient == nil {
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

	_, appErr = h.natsClient.Subscribe(dto.JobEmailVerification, func(b []byte) {
		var data dto.EmailVerificationDTO
		err := json.Unmarshal(b, &data)
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

		appErr = mailer.Send(client, mailer.Message{
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
	})
	if appErr != nil {
		return appErr
	}

	_, appErr = h.natsClient.Subscribe(dto.JobEmailPasswordReset, func(b []byte) {
		var data dto.PasswordResetDTO
		err := json.Unmarshal(b, &data)
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

		appErr = mailer.Send(client, mailer.Message{
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
	})
	if appErr != nil {
		return appErr
	}

	return nil
}
