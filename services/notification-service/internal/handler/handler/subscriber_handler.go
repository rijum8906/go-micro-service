// Package broker
package handler

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/mailer"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/internal/constants"
	"github.com/rijum8906/relay/services/notification-service/internal/services/subscriber"
)

type SubscribeHandler struct {
	SubscriberService subscriber.Service
	BrokerClient      broker.Client
	templateManager   template.TemplateManager
	mailerCfg         *mailer.Config
}

type ConsumerConfig struct {
	Stream string
	Config *nats.ConsumerConfig
}

func New(subService subscriber.Service, client broker.Client, mailerCfg *mailer.Config, tm template.TemplateManager) (*SubscribeHandler, *apperror.AppError) {
	return &SubscribeHandler{
		SubscriberService: subService,
		BrokerClient:      client,
		mailerCfg:         mailerCfg,
		templateManager:   tm,
	}, nil
}

func (h *SubscribeHandler) CreateStreams() *apperror.AppError {
	streamManager := broker.NewStreamManager(h.BrokerClient.GetClient())

	// Verification Stream Config
	config := broker.NewStreamConfig(constants.StreamVerification).
		AddSubjects(string(dto.JobEmailVerification), string(dto.JobEmailPasswordReset))
	_, appErr := streamManager.Create(config)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (h *SubscribeHandler) CreateConsumers() *apperror.AppError {
	consumerManager := broker.NewConsumerManager(h.BrokerClient.GetClient())

	// Verification Comsumer
	config := broker.NewConsumerConfig(constants.ConsumerVerification)

	_, appErr := consumerManager.Create(constants.StreamVerification, config)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (h *SubscribeHandler) Subscribe() *apperror.AppError {
	if h == nil {
		return apperror.ErrInternal.WithMessage("subscribe handler is not initialized")
	}

	if h.BrokerClient == nil {
		return apperror.ErrInternal.WithMessage("nats client is not initialized")
	}

	if h.mailerCfg == nil {
		return apperror.ErrInternal.WithMessage("mailer config is not initialized")
	}

	if h.templateManager == nil {
		return apperror.ErrInternal.WithMessage("template manager is not initialized")
	}

	go func(service subscriber.Service) {
		if appErr := service.SubscribeEmailVerificationJob(constants.ConsumerVerification); appErr != nil {
			fmt.Println(appErr.Details)
		}
	}(h.SubscriberService)
	go func(service subscriber.Service) {
		if appErr := service.SubscribeJobPasswordReset(constants.ConsumerVerification); appErr != nil {
			fmt.Println(appErr.Details)
		}
	}(h.SubscriberService)

	return nil
}
