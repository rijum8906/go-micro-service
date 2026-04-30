// Package userauth
package userauth

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/jobs"
	"github.com/rijum8906/relay/services/notification-service/internal/constants"
	"go.uber.org/zap"
)

// CreateConsumer creates a consumer for the user auth email jobs
func (s *UserAuthEmailService) CreateConsumer() *apperror.AppError {
	consumerManager := broker.NewConsumerManager(s.BrokerClient.GetClient())

	config := broker.NewConsumerConfig(constants.ConsumerUserAuth).AddDeliverPolicy(nats.DeliverAllPolicy).
		WithFilterSubject(jobs.JobUserAuthWildcard).
		AddMaxDelivery(3)
	// TODO: make the config with env variables

	exists, appErr := consumerManager.Exists(constants.StreamUser, constants.ConsumerUserAuth)
	if appErr != nil {
		return appErr
	}
	if exists {
		consumerInfo, appErr := consumerManager.Update(constants.StreamUser, config)
		if appErr != nil {
			return appErr
		}
		if consumerInfo != nil {
			s.ConsumerInfo = consumerInfo
			return nil
		}

		return nil
	}

	info, appErr := consumerManager.Create(constants.StreamUser, config)
	if appErr != nil {
		return appErr
	}

	s.ConsumerInfo = info

	return nil
}

func (s *UserAuthEmailService) ListenMessage() {
	subscriber := broker.NewSubscriber(s.BrokerClient.GetClient())

	subscription, appErr := subscriber.PullSubscribe(jobs.JobUserAuthWildcard, constants.ConsumerUserAuth)
	if appErr != nil {
		s.Logger.Error("failed to subscribe", zap.String("error_message", appErr.Message), zap.Any("details", appErr.Details))
	}

	batchSize := 5
	waitTime := 30 * time.Second

	s.Logger.Info("started listening messages", zap.String("consumer_name", s.ConsumerInfo.Name))

	for {
		msgs, err := subscription.Fetch(batchSize, nats.MaxWait(waitTime))
		if err != nil {
			if err == nats.ErrTimeout {
				time.Sleep(1 * time.Second)
				continue
			}
			s.Logger.Error("failed to fetch messages", zap.Error(err))
		}

		for _, msg := range msgs {
			s.processMessage(msg)
		}
	}
}

func (s *UserAuthEmailService) processMessage(msg *nats.Msg) {
	// Check retry count
	metadata, err := msg.Metadata()
	if err == nil && metadata.NumDelivered > uint64(s.ConsumerInfo.Config.MaxDeliver) {
		s.Logger.Error("max retries exceeded, discarding",
			zap.String("subject", msg.Subject),
			zap.Int("retries", int(metadata.NumDelivered)))

		_ = msg.Term()
		return
	}

	var processErr *apperror.AppError
	// Process message
	switch msg.Subject {
	case jobs.JobUserRequestedPasswordReset:
		processErr = s.processPasswordReset(msg)
	case jobs.JobUserRequestedEmailVerification:
		processErr = s.processEmailVerification(msg)
	default:
		s.Logger.Warn("unknown job subject", zap.String("subject", msg.Subject))
		_ = msg.Ack() // IMPORTANT: Ack unknown to avoid infinite retries
	}

	if processErr != nil {
		s.Logger.Error("error processing message",
			zap.String("subject", msg.Subject),
			zap.String("error_message", processErr.Message),
			zap.Any("details", processErr.Details))

		// TODO: add delay
		_ = msg.Nak() // Negative ack, will retry
		return
	}

	// Ack successful messages
	_ = msg.Ack()
}
