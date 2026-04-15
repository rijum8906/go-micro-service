// Package broker
package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/nats"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/internal/repository/notificationlog"
	"github.com/rijum8906/relay/services/notification-service/internal/services/email"
	"github.com/rijum8906/relay/services/notification-service/internal/utils"
	"github.com/wneessen/go-mail"
)

type SubscribeHandler struct {
	EmailService        email.Service
	NatsClient          *nats.Client
	templateManager     template.TemplateManager
	mailerCfg           *mailer.Config
	mailerClient        *mail.Client
	notificationLogRepo notificationlog.NotificationLogRepository
}

func New(emailService email.Service, client *nats.Client, mailerCfg *mailer.Config, notificationLogRepo notificationlog.NotificationLogRepository) (*SubscribeHandler, *apperror.AppError) {
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
		EmailService:        emailService,
		NatsClient:          client,
		mailerCfg:           mailerCfg,
		templateManager:     tm,
		notificationLogRepo: notificationLogRepo,
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

	if h.notificationLogRepo == nil {
		return apperror.ErrInternal.WithMessage("notification log repository is not initialized")
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
	ctx := context.Background()
	jobID, appErr := h.notificationLogRepo.CreateJob(ctx, notificationlog.CreateJobParams{
		JobSubject: string(dto.JobEmailVerification),
		RawPayload: string(raw),
	})
	if appErr != nil {
		fmt.Println("Error creating notification job log:", appErr.Details)
		return
	}
	if appErr = h.notificationLogRepo.MarkJobProcessing(ctx, jobID); appErr != nil {
		fmt.Println("Error marking notification job processing:", appErr.Details)
	}

	var data dto.EmailVerificationDTO
	err := json.Unmarshal(raw, &data)
	if err != nil {
		if appErr = h.notificationLogRepo.MarkJobInvalidPayload(ctx, jobID, err.Error()); appErr != nil {
			fmt.Println("Error marking notification job invalid payload:", appErr.Details)
		}
		fmt.Println("Error unmarshalling job:", err)
		return
	}

	deliveryID, appErr := h.notificationLogRepo.CreateDelivery(ctx, notificationlog.CreateDeliveryParams{
		JobID:            jobID,
		Channel:          "email",
		NotificationType: "email_verification",
		TemplateType:     string(template.TemplateTypeEmailVerification),
		RecipientEmail:   data.ClientEmail,
		RecipientName:    data.ClientName,
		Subject:          "Email Verification",
		Provider:         "smtp",
	})
	if appErr != nil {
		_ = h.notificationLogRepo.MarkJobFailed(ctx, jobID, appErrorMessage(appErr))
		fmt.Println("Error creating notification delivery log:", appErr.Details)
		return
	}
	if appErr = h.notificationLogRepo.MarkDeliverySending(ctx, deliveryID); appErr != nil {
		_ = h.notificationLogRepo.MarkJobFailed(ctx, jobID, appErrorMessage(appErr))
		fmt.Println("Error marking notification delivery sending:", appErr.Details)
		return
	}

	emailTemplate, appErr := h.templateManager.RenderToString(template.TemplateTypeEmailVerification, data)
	if appErr != nil {
		h.logFailedAttempt(ctx, jobID, deliveryID, appErrorMessage(appErr))
		fmt.Println("Error rendering email template:", appErr.Details)
		return
	}

	envelop, appErr := utils.ParseMailEnvelop(h.mailerCfg, data.ClientEmail)
	if appErr != nil {
		h.logFailedAttempt(ctx, jobID, deliveryID, appErrorMessage(appErr))
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
		h.logFailedAttempt(ctx, jobID, deliveryID, appErrorMessage(appErr))
		fmt.Println("Error sending email:", appErr.Details)
		return
	}

	if _, appErr = h.notificationLogRepo.CreateAttemptSuccess(ctx, notificationlog.CreateAttemptParams{
		DeliveryID: deliveryID,
		AttemptNo:  1,
		Provider:   "smtp",
	}); appErr != nil {
		fmt.Println("Error creating notification success attempt log:", appErr.Details)
	}
	if appErr = h.notificationLogRepo.MarkDeliverySent(ctx, deliveryID, "smtp", ""); appErr != nil {
		fmt.Println("Error marking notification delivery sent:", appErr.Details)
		return
	}
	if appErr = h.notificationLogRepo.MarkJobCompleted(ctx, jobID); appErr != nil {
		fmt.Println("Error marking notification job completed:", appErr.Details)
		return
	}
}

func (h *SubscribeHandler) handlerPasswordReset(raw []byte) {
	ctx := context.Background()
	jobID, appErr := h.notificationLogRepo.CreateJob(ctx, notificationlog.CreateJobParams{
		JobSubject: string(dto.JobEmailPasswordReset),
		RawPayload: string(raw),
	})
	if appErr != nil {
		fmt.Println("Error creating notification job log:", appErr.Details)
		return
	}
	if appErr = h.notificationLogRepo.MarkJobProcessing(ctx, jobID); appErr != nil {
		fmt.Println("Error marking notification job processing:", appErr.Details)
	}

	var data dto.PasswordResetDTO
	err := json.Unmarshal(raw, &data)
	if err != nil {
		if appErr = h.notificationLogRepo.MarkJobInvalidPayload(ctx, jobID, err.Error()); appErr != nil {
			fmt.Println("Error marking notification job invalid payload:", appErr.Details)
		}
		fmt.Println("Error unmarshalling job:", err)
		return
	}

	deliveryID, appErr := h.notificationLogRepo.CreateDelivery(ctx, notificationlog.CreateDeliveryParams{
		JobID:            jobID,
		Channel:          "email",
		NotificationType: "password_reset",
		TemplateType:     string(template.TemplateTypeEmailPasswordReset),
		RecipientEmail:   data.ClientEmail,
		RecipientName:    data.ClientName,
		Subject:          "Password Reset",
		Provider:         "smtp",
	})
	if appErr != nil {
		_ = h.notificationLogRepo.MarkJobFailed(ctx, jobID, appErrorMessage(appErr))
		fmt.Println("Error creating notification delivery log:", appErr.Details)
		return
	}
	if appErr = h.notificationLogRepo.MarkDeliverySending(ctx, deliveryID); appErr != nil {
		_ = h.notificationLogRepo.MarkJobFailed(ctx, jobID, appErrorMessage(appErr))
		fmt.Println("Error marking notification delivery sending:", appErr.Details)
		return
	}

	emailTemplate, appErr := h.templateManager.RenderToString(template.TemplateTypeEmailPasswordReset, data)
	if appErr != nil {
		h.logFailedAttempt(ctx, jobID, deliveryID, appErrorMessage(appErr))
		fmt.Println("Error rendering email template:", appErr.Details)
		return
	}

	envelop, appErr := utils.ParseMailEnvelop(h.mailerCfg, data.ClientEmail)
	if appErr != nil {
		h.logFailedAttempt(ctx, jobID, deliveryID, appErrorMessage(appErr))
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
		h.logFailedAttempt(ctx, jobID, deliveryID, appErrorMessage(appErr))
		fmt.Println("Error sending email:", appErr.Details)
		return
	}

	if _, appErr = h.notificationLogRepo.CreateAttemptSuccess(ctx, notificationlog.CreateAttemptParams{
		DeliveryID: deliveryID,
		AttemptNo:  1,
		Provider:   "smtp",
	}); appErr != nil {
		fmt.Println("Error creating notification success attempt log:", appErr.Details)
	}
	if appErr = h.notificationLogRepo.MarkDeliverySent(ctx, deliveryID, "smtp", ""); appErr != nil {
		fmt.Println("Error marking notification delivery sent:", appErr.Details)
		return
	}
	if appErr = h.notificationLogRepo.MarkJobCompleted(ctx, jobID); appErr != nil {
		fmt.Println("Error marking notification job completed:", appErr.Details)
		return
	}
}

func (h *SubscribeHandler) logFailedAttempt(ctx context.Context, jobID, deliveryID uuid.UUID, errMessage string) {
	if _, appErr := h.notificationLogRepo.CreateAttemptFailed(ctx, notificationlog.CreateAttemptParams{
		DeliveryID:   deliveryID,
		AttemptNo:    1,
		ErrorMessage: errMessage,
		ErrorDetails: "{}",
		Provider:     "smtp",
	}); appErr != nil {
		fmt.Println("Error creating notification failed attempt log:", appErr.Details)
	}
	if appErr := h.notificationLogRepo.MarkDeliveryFailed(ctx, deliveryID, errMessage); appErr != nil {
		fmt.Println("Error marking notification delivery failed:", appErr.Details)
	}
	if appErr := h.notificationLogRepo.MarkJobFailed(ctx, jobID, errMessage); appErr != nil {
		fmt.Println("Error marking notification job failed:", appErr.Details)
	}
}

func appErrorMessage(appErr *apperror.AppError) string {
	if appErr == nil {
		return ""
	}
	if len(appErr.Details) == 0 {
		return appErr.Message
	}

	return fmt.Sprintf("%s: %v", appErr.Message, appErr.Details)
}
