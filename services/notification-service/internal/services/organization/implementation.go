// Package organizationservice
package organizationservice

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/jobs"
	"github.com/rijum8906/relay/services/notification-service/internal/constants"
	orgauth "github.com/rijum8906/relay/services/notification-service/internal/services/organization/auth"
	"github.com/rijum8906/relay/services/notification-service/internal/services/organization/openfga"
)

func (s *OrganizationService) CreateStream() *apperror.AppError {
	streamManager := broker.NewStreamManager(s.BrokerClient.GetClient())

	config := broker.NewStreamConfig(constants.StreamOrganization).
		AddStorageType(nats.FileStorage).
		AddSubjects(jobs.GetDomainWildcard(jobs.JobOrganizationCreated)).
		AddMaxConsumer(5)

	exists, appErr := streamManager.Exists(constants.StreamOrganization)
	if appErr != nil {
		return appErr
	}
	if exists {
		streamInfo, appErr := streamManager.Update(config)
		if appErr != nil {
			return appErr
		}
		if streamInfo != nil {
			s.StreamInfo = streamInfo
			return nil
		}

		return nil
	}

	_, appErr = streamManager.Create(config)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (s *OrganizationService) Run() *apperror.AppError {
	// Auth Service
	authService, appErr := orgauth.New()
	if appErr != nil {
		return appErr
	}

	if appErr = authService.CreateConsumer(); appErr != nil {
		return appErr
	}

	go func(service *orgauth.OrgAuthEmailService) {
		service.ListenMessage()
	}(authService)

	// OpenFGA Service
	openfgaService, appErr := openfga.New()
	if appErr != nil {
		return appErr
	}

	if appErr = openfgaService.CreateConsumer(); appErr != nil {
		return appErr
	}

	if appErr = openfgaService.InitOpenFGA(); appErr != nil {
		return appErr
	}

	go func(service *openfga.OrgPermissionService) {
		service.ListenMessage()
	}(openfgaService)

	return nil
}
