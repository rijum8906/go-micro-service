package notification

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreutils"
	"github.com/rijum8906/relay/services/notification-service/internal/db"
)

func (s *service) CreateNotification(ctx context.Context, params db.CreateNotificationParams) (*db.Notification, *apperror.AppError) {
	notif, err := s.q.CreateNotification(ctx, params)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}
	return &notif, nil
}

func (s *service) GetNotification(ctx context.Context, id string) (*db.Notification, *apperror.AppError) {
	uuid, appErr := coreutils.ParseToUUID(id)
	if appErr != nil {
		return nil, appErr
	}

	notif, err := s.q.GetNotification(ctx, uuid)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	return &notif, nil
}

func (s *service) GetNotificationsByUserID(ctx context.Context, userID string, limit, page int) (*[]db.Notification, *apperror.AppError) {
	userUUID, appErr := coreutils.ParseToUUID(userID)
	if appErr != nil {
		return nil, appErr
	}

	offset := (page - 1) * limit
	notifications, err := s.q.GetNotificationsByUserID(ctx, db.GetNotificationsByUserIDParams{
		RecepientUserID: userUUID,
		Limit:           int32(limit),
		Offset:          int32(offset),
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("error", err.Error())
	}

	return &notifications, nil
}

func (s *service) UpdateNotificationStatus(ctx context.Context, params db.UpdateNotificationStatusParams) *apperror.AppError {
	err := s.q.UpdateNotificationStatus(ctx, params)
	if err != nil {
		return apperror.ErrInternal.WithDetail("error", err.Error())
	}
	return nil
}

func (s *service) DeleteNotification(ctx context.Context, id string) *apperror.AppError {
	uuid, appErr := coreutils.ParseToUUID(id)
	if appErr != nil {
		return appErr
	}

	err := s.q.DeleteNotification(ctx, uuid)
	if err != nil {
		return apperror.ErrInternal.WithDetail("error", err.Error())
	}

	return nil
}
