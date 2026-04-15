package notification

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/notification-service/internal/db"
)

func (s *service) CreateNotification(ctx context.Context, params db.CreateNotificationParams) (*db.Notification, *apperror.AppError) {
	return nil, nil
}

func (s *service) GetNotification(ctx context.Context, id string) (*db.Notification, *apperror.AppError) {
	return nil, nil
}

func (s *service) GetNotificationsByUserID(ctx context.Context, userID string) (*[]db.Notification, *apperror.AppError) {
	return nil, nil
}

func (s *service) UpdateNotificationStatus(ctx context.Context, params db.UpdateNotificationStatusParams) *apperror.AppError {
	return nil
}

func (s *service) DeleteNotification(ctx context.Context, id string) *apperror.AppError {
	return nil
}
