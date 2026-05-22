// Package openfga
package openfga

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/app"
	"github.com/rijum8906/relay/services/notification-service/app/config"
	"go.uber.org/zap"
)

type OrgPermissionService struct {
	// Core
	BrokerClient    broker.Client
	TemplateManager template.TemplateManager

	// Utilities
	Logger            *zap.Logger
	FgaClient         *coreopenfga.Client
	TupleManager      coreopenfga.TuppleManager
	PermissionManager *permissions.PermissionManager

	// Config
	ConsumerInfo *nats.ConsumerInfo
	AppConfig    *config.Env
}

func New() (*OrgPermissionService, *apperror.AppError) {
	application, appErr := app.GetInstance()
	if appErr != nil {
		return nil, appErr
	}

	return &OrgPermissionService{
		BrokerClient:    application.BrokerClient(),
		TemplateManager: application.TemplateManager(),
		Logger:          application.Logger(),
		AppConfig:       application.Config(),
	}, nil
}
