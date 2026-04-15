// Package notification
package notification

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/notification-service/internal/db"
)

type Service interface {
	CreateNotification(ctx context.Context, params db.CreateNotificationParams) (*db.Notification, *apperror.AppError)
	GetNotification(ctx context.Context, id string) (*db.Notification, *apperror.AppError)
	GetNotificationsByUserID(ctx context.Context, userID string) (*[]db.Notification, *apperror.AppError)
	UpdateNotificationStatus(ctx context.Context, params db.UpdateNotificationStatusParams) *apperror.AppError
	DeleteNotification(ctx context.Context, id string) *apperror.AppError
}

type service struct {
	q db.Querier
}

func New() Service {
	return &service{}
}
