// Package userservice
package userservice

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/jobs"
	"github.com/rijum8906/relay/services/notification-service/internal/constants"
	userauth "github.com/rijum8906/relay/services/notification-service/internal/services/user/auth"
)

func (s *UserService) CreateStream() *apperror.AppError {
	streamManager := broker.NewStreamManager(s.BrokerClient.GetClient())

	config := broker.NewStreamConfig(constants.StreamUser).
		AddStorageType(nats.FileStorage).
		AddSubjects(jobs.GetDomainWildcard(jobs.JobUserRequestedEmailVerification))

	exists, appErr := streamManager.Exists(constants.StreamUser)
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

func (s *UserService) Run() *apperror.AppError {
	authService, appErr := userauth.New()
	if appErr != nil {
		return appErr
	}

	if appErr = authService.CreateConsumer(); appErr != nil {
		return appErr
	}

	go func(service *userauth.UserAuthEmailService) {
		service.ListenMessage()
	}(authService)

	return nil
}
