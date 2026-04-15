package notificationlog

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type notificationLogRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) NotificationLogRepository {
	return &notificationLogRepository{
		db: db,
	}
}

func (r *notificationLogRepository) CreateJob(ctx context.Context, params CreateJobParams) (uuid.UUID, *apperror.AppError) {
	db, appErr := r.database()
	if appErr != nil {
		return uuid.Nil, appErr
	}

	var id uuid.UUID
	err := db.QueryRow(ctx, `
		INSERT INTO notification_jobs (job_subject, raw_payload)
		VALUES ($1, $2)
		RETURNING id
	`, params.JobSubject, params.RawPayload).Scan(&id)
	if err != nil {
		return uuid.Nil, databaseError("failed to create notification job", err)
	}

	return id, nil
}

func (r *notificationLogRepository) MarkJobProcessing(ctx context.Context, id uuid.UUID) *apperror.AppError {
	return r.execOne(ctx, "notification job not found", "failed to mark notification job processing", `
		UPDATE notification_jobs
		SET status = 'processing', updated_at = now()
		WHERE id = $1
	`, id)
}

func (r *notificationLogRepository) MarkJobCompleted(ctx context.Context, id uuid.UUID) *apperror.AppError {
	return r.execOne(ctx, "notification job not found", "failed to mark notification job completed", `
		UPDATE notification_jobs
		SET status = 'completed', completed_at = now(), updated_at = now()
		WHERE id = $1
	`, id)
}

func (r *notificationLogRepository) MarkJobFailed(ctx context.Context, id uuid.UUID, errorMessage string) *apperror.AppError {
	return r.execOne(ctx, "notification job not found", "failed to mark notification job failed", `
		UPDATE notification_jobs
		SET status = 'failed', error_message = $2, completed_at = now(), updated_at = now()
		WHERE id = $1
	`, id, errorMessage)
}

func (r *notificationLogRepository) MarkJobInvalidPayload(ctx context.Context, id uuid.UUID, errorMessage string) *apperror.AppError {
	return r.execOne(ctx, "notification job not found", "failed to mark notification job invalid payload", `
		UPDATE notification_jobs
		SET status = 'invalid_payload', error_message = $2, completed_at = now(), updated_at = now()
		WHERE id = $1
	`, id, errorMessage)
}

func (r *notificationLogRepository) CreateDelivery(ctx context.Context, params CreateDeliveryParams) (uuid.UUID, *apperror.AppError) {
	db, appErr := r.database()
	if appErr != nil {
		return uuid.Nil, appErr
	}

	var id uuid.UUID
	err := db.QueryRow(ctx, `
		INSERT INTO notification_deliveries (
			job_id,
			channel,
			notification_type,
			template_type,
			recipient_email,
			recipient_name,
			subject,
			provider
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`,
		params.JobID,
		valueOrDefault(params.Channel, "email"),
		params.NotificationType,
		nullIfEmpty(params.TemplateType),
		nullIfEmpty(params.RecipientEmail),
		nullIfEmpty(params.RecipientName),
		nullIfEmpty(params.Subject),
		nullIfEmpty(params.Provider),
	).Scan(&id)
	if err != nil {
		return uuid.Nil, databaseError("failed to create notification delivery", err)
	}

	return id, nil
}

func (r *notificationLogRepository) MarkDeliverySending(ctx context.Context, id uuid.UUID) *apperror.AppError {
	return r.execOne(ctx, "notification delivery not found", "failed to mark notification delivery sending", `
		UPDATE notification_deliveries
		SET status = 'sending', updated_at = now()
		WHERE id = $1
	`, id)
}

func (r *notificationLogRepository) MarkDeliverySent(ctx context.Context, id uuid.UUID, provider, providerMessageID string) *apperror.AppError {
	return r.execOne(ctx, "notification delivery not found", "failed to mark notification delivery sent", `
		UPDATE notification_deliveries
		SET status = 'sent',
			provider = COALESCE($2, provider),
			provider_message_id = COALESCE($3, provider_message_id),
			sent_at = now(),
			updated_at = now()
		WHERE id = $1
	`, id, nullIfEmpty(provider), nullIfEmpty(providerMessageID))
}

func (r *notificationLogRepository) MarkDeliveryFailed(ctx context.Context, id uuid.UUID, errorMessage string) *apperror.AppError {
	return r.execOne(ctx, "notification delivery not found", "failed to mark notification delivery failed", `
		UPDATE notification_deliveries
		SET status = 'failed', last_error = $2, failed_at = now(), updated_at = now()
		WHERE id = $1
	`, id, errorMessage)
}

func (r *notificationLogRepository) CreateAttemptSuccess(ctx context.Context, params CreateAttemptParams) (uuid.UUID, *apperror.AppError) {
	return r.createAttempt(ctx, "success", params)
}

func (r *notificationLogRepository) CreateAttemptFailed(ctx context.Context, params CreateAttemptParams) (uuid.UUID, *apperror.AppError) {
	return r.createAttempt(ctx, "failed", params)
}

func (r *notificationLogRepository) createAttempt(ctx context.Context, status string, params CreateAttemptParams) (uuid.UUID, *apperror.AppError) {
	db, appErr := r.database()
	if appErr != nil {
		return uuid.Nil, appErr
	}

	var id uuid.UUID
	err := db.QueryRow(ctx, `
		INSERT INTO notification_delivery_attempts (
			delivery_id,
			attempt_no,
			status,
			error_message,
			error_details,
			provider,
			provider_message_id,
			finished_at
		)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, now())
		RETURNING id
	`,
		params.DeliveryID,
		params.AttemptNo,
		status,
		nullIfEmpty(params.ErrorMessage),
		valueOrDefault(params.ErrorDetails, "{}"),
		nullIfEmpty(params.Provider),
		nullIfEmpty(params.ProviderMessageID),
	).Scan(&id)
	if err != nil {
		return uuid.Nil, databaseError("failed to create notification delivery attempt", err)
	}

	return id, nil
}

func (r *notificationLogRepository) database() (*pgxpool.Pool, *apperror.AppError) {
	if r == nil || r.db == nil {
		return nil, apperror.New(apperror.CodeInternal, "notification log repository is not initialized")
	}

	return r.db, nil
}

func (r *notificationLogRepository) execOne(ctx context.Context, notFoundMessage, errorMessage, query string, args ...any) *apperror.AppError {
	db, appErr := r.database()
	if appErr != nil {
		return appErr
	}

	tag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return databaseError(errorMessage, err)
	}
	if tag.RowsAffected() == 0 {
		return apperror.New(apperror.CodeNotFound, notFoundMessage)
	}

	return nil
}

func databaseError(message string, err error) *apperror.AppError {
	return apperror.New(apperror.CodeDatabase, message).WithDetail("error", err.Error())
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return value
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

var _ NotificationLogRepository = (*notificationLogRepository)(nil)
