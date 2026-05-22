package organizationservice

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/services/notification-service/app"
	"go.uber.org/zap"
)

type OrganizationService struct {
	// core
	BrokerClient broker.Client

	// utilities
	Logger *zap.Logger

	// config
	StreamInfo *nats.StreamInfo
}

func New() (*OrganizationService, *apperror.AppError) {
	application, appErr := app.GetInstance()
	if appErr != nil {
		return nil, appErr
	}

	return &OrganizationService{
		BrokerClient: application.BrokerClient(),
		Logger:       application.Logger(),
	}, nil
}
