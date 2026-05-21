package userservice

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/services/notification-service/app"
	"go.uber.org/zap"
)

type UserService struct {
	// Core
	BrokerClient broker.Client

	// Utilities
	Logger *zap.Logger
}

func New() (*UserService, *apperror.AppError) {
	application, appErr := app.GetInstance()
	if appErr != nil {
		return nil, appErr
	}

	return &UserService{
		BrokerClient: application.BrokerClient(),
		Logger:       application.Logger(),
	}, nil
}
