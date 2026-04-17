// Package broker
package broker

import (
	"time"

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
	NatsClient        *broker.Client
	templateManager   template.TemplateManager
	mailerCfg         *mailer.Config
}

type ConsumerConfig struct {
	Stream string
	Config *nats.ConsumerConfig
}

type Stream struct {
	Stream  constants.StreamName
	Subject dto.JobSubject
}

func New(subService subscriber.Service, client *broker.Client, mailerCfg *mailer.Config, tm template.TemplateManager) (*SubscribeHandler, *apperror.AppError) {
	return &SubscribeHandler{
		SubscriberService: subService,
		NatsClient:        client,
		mailerCfg:         mailerCfg,
		templateManager:   tm,
	}, nil
}

func (h *SubscribeHandler) CreateStreams() *apperror.AppError {
	streams := []*Stream{
		{
			Stream:  constants.StreamEmailVerification,
			Subject: dto.JobEmailVerification,
		}, {
			Stream:  constants.StreamPasswordReset,
			Subject: dto.JobEmailPasswordReset,
		},
	}

	for _, stream := range streams {
		err := h.NatsClient.EnsureStream(broker.StreamConfig{
			Name:       string(stream.Stream),
			Subjects:   []string{string(stream.Subject)},
			MaxBytes:   1024 * 1024,
			MaxMsgs:    1000,
			MaxAge:     time.Hour,
			MaxMsgSize: 1024 * 1024,
			Storage:    nats.FileStorage,
			Replicas:   1,
			Duplicates: time.Minute,
			Discard:    nats.DiscardNew,
		})
		if err != nil {
			return apperror.ErrInternal.WithMessage("failed to create stream").WithDetail("error", err.Error())
		}
	}

	return nil
}

func (h *SubscribeHandler) CreateConsumers() *apperror.AppError {
	consumersCfgs := []*ConsumerConfig{
		{
			Stream: string(constants.StreamEmailVerification),
			Config: &nats.ConsumerConfig{
				Durable:       string(constants.DurableEmailVerification),
				Name:          string(constants.StreamEmailVerification),
				DeliverPolicy: nats.DeliverAllPolicy,
				AckPolicy:     nats.AckAllPolicy,
				ReplayPolicy:  nats.ReplayOriginalPolicy,
			},
		},
		{
			Stream: string(constants.StreamPasswordReset),
			Config: &nats.ConsumerConfig{
				Durable:       string(constants.DurablePasswordReset),
				Name:          string(constants.StreamPasswordReset),
				DeliverPolicy: nats.DeliverAllPolicy,
				AckPolicy:     nats.AckAllPolicy,
				ReplayPolicy:  nats.ReplayOriginalPolicy,
			},
		},
	}

	for _, cfg := range consumersCfgs {
		_, err := h.NatsClient.JS.AddConsumer(cfg.Stream, cfg.Config)
		if err != nil {
			return apperror.ErrInternal.WithMessage("error adding consumer").WithDetail("error", err.Error())
		}
	}

	return nil
}

func (h *SubscribeHandler) Subscribe() *apperror.AppError {
	if h == nil {
		return apperror.ErrInternal.WithMessage("subscribe handler is not initialized")
	}

	if h.NatsClient == nil {
		return apperror.ErrInternal.WithMessage("nats client is not initialized")
	}

	if h.mailerCfg == nil {
		return apperror.ErrInternal.WithMessage("mailer config is not initialized")
	}

	if h.templateManager == nil {
		return apperror.ErrInternal.WithMessage("template manager is not initialized")
	}

	h.SubscriberService.SubscribeEmailVerificationJob(constants.DurableEmailVerification, constants.StreamEmailVerification)

	return nil
}
