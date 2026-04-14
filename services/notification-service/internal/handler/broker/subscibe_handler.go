// Package broker
package broker

import (
	"encoding/json"

	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/nats"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/internal/services/email"
	"github.com/rijum8906/relay/services/notification-service/internal/utils"
)

type SubscribeHandler struct {
	EmailService    email.Service
	NatsClient      *nats.Client
	templateManager template.TemplateManager
	mailerCfg       *mailer.Config
}

func New(emailService email.Service, client *nats.Client) *SubscribeHandler {
	return &SubscribeHandler{
		EmailService: emailService,
		NatsClient:   client,
	}
}

func (h *SubscribeHandler) Subscribe() {
	client, appErr := mailer.Connect(*h.mailerCfg)
	if appErr != nil {
		panic(appErr)
	}

	h.NatsClient.Subscribe(dto.JobEmailVerification, func(b []byte) {
		var data dto.EmailVerificationJob
		err := json.Unmarshal(b, &data)
		if err != nil {
			// TODO: save to some logs
			return
		}

		emailTemplate, appErr := h.templateManager.RenderToString(template.TemplateTypeEmailVerification, data)
		if appErr != nil {
			// TODO: save to some logs
			return
		}

		envelop, appErr := utils.ParseMailEnvelop(h.mailerCfg, data.Email)
		if appErr != nil {
			// TODO: save to some logs
			return
		}

		mailer.Send(client, mailer.Message{
			Envelope: envelop,
			Content: mailer.Content{
				HTML:    emailTemplate,
				Subject: "Email Verification",
			},
		})
	})
}
