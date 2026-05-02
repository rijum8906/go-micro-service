// Package handler
package handler

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/jobs"
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
	// Stream User Auth
	subjectWildcard := jobs.GetSubdomainWildcard(jobs.JobUserRequestedEmailVerification)
	streamCfg := broker.NewStreamConfig(constants.StreamUser).
		AddSubjects(subjectWildcard)
	userAuthStream := broker.NewStreamManager(h.BrokerClient.GetClient())
	_, appErr := userAuthStream.Create(streamCfg)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (h *SubscribeHandler) CreateConsumers() *apperror.AppError {
	// Consumer Request Emails
	subjectWildcard := jobs.GetSubdomainWildcard(jobs.JobUserRequestedEmailVerification)
	consumerCfg := broker.NewConsumerConfig(constants.ConsumerUserAuth).AddDeliverPolicy(nats.DeliverAllPolicy).
		WithFilterSubject(subjectWildcard)
	consumerHandler := broker.NewConsumerManager(h.BrokerClient.GetClient())
	_, appErr := consumerHandler.Create(constants.StreamUser, consumerCfg)
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
		if appErr := service.SubscribeUserAuthEmailJobs(constants.ConsumerUserAuth); appErr != nil {
			fmt.Println(appErr.Details)
		}
	}(h.SubscriberService)
	return nil
}
