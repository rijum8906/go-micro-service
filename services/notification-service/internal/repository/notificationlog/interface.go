package notificationlog

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type CreateJobParams struct {
	JobSubject string
	RawPayload string
}

type CreateDeliveryParams struct {
	JobID            uuid.UUID
	Channel          string
	NotificationType string
	TemplateType     string
	RecipientEmail   string
	RecipientName    string
	Subject          string
	Provider         string
}

type CreateAttemptParams struct {
	DeliveryID        uuid.UUID
	AttemptNo         int
	ErrorMessage      string
	ErrorDetails      string
	Provider          string
	ProviderMessageID string
}

type NotificationLogRepository interface {
	CreateJob(ctx context.Context, params CreateJobParams) (uuid.UUID, *apperror.AppError)
	MarkJobProcessing(ctx context.Context, id uuid.UUID) *apperror.AppError
	MarkJobCompleted(ctx context.Context, id uuid.UUID) *apperror.AppError
	MarkJobFailed(ctx context.Context, id uuid.UUID, errorMessage string) *apperror.AppError
	MarkJobInvalidPayload(ctx context.Context, id uuid.UUID, errorMessage string) *apperror.AppError

	CreateDelivery(ctx context.Context, params CreateDeliveryParams) (uuid.UUID, *apperror.AppError)
	MarkDeliverySending(ctx context.Context, id uuid.UUID) *apperror.AppError
	MarkDeliverySent(ctx context.Context, id uuid.UUID, provider, providerMessageID string) *apperror.AppError
	MarkDeliveryFailed(ctx context.Context, id uuid.UUID, errorMessage string) *apperror.AppError

	CreateAttemptSuccess(ctx context.Context, params CreateAttemptParams) (uuid.UUID, *apperror.AppError)
	CreateAttemptFailed(ctx context.Context, params CreateAttemptParams) (uuid.UUID, *apperror.AppError)
}
