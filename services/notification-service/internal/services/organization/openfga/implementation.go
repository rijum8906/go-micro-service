package openfga

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/broker"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/jobs"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/services/notification-service/internal/constants"
	"go.uber.org/zap"
)

// CreateConsumer creates a consumer for the user auth email jobs
func (s *OrgPermissionService) CreateConsumer() *apperror.AppError {
	consumerManager := broker.NewConsumerManager(s.BrokerClient.GetClient())

	config := broker.NewConsumerConfig(constants.ConsumerOrganizationOpenFGA).AddDeliverPolicy(nats.DeliverAllPolicy).
		WithFilterSubject(jobs.JobOrganizationRoleWildcard).
		AddMaxDelivery(3)
	// TODO: make the config with env variables

	exists, appErr := consumerManager.Exists(constants.StreamOrganization, constants.ConsumerOrganizationOpenFGA)
	if appErr != nil {
		return appErr
	}
	if exists {
		consumerInfo, appErr := consumerManager.Update(constants.StreamOrganization, config)
		if appErr != nil {
			return appErr
		}
		if consumerInfo != nil {
			s.ConsumerInfo = consumerInfo
			return nil
		}

		return nil
	}

	info, appErr := consumerManager.Create(constants.StreamOrganization, config)
	if appErr != nil {
		return appErr
	}

	s.ConsumerInfo = info

	return nil
}

func (s *OrgPermissionService) InitOpenFGA() *apperror.AppError {
	client, appErr := coreopenfga.NewClient(s.AppConfig.FGAAPIURL)
	if appErr != nil {
		return appErr
	}
	s.FgaClient = client

	tupleManager := coreopenfga.NewTupleManager(s.FgaClient)
	s.TupleManager = tupleManager

	permissionManager := permissions.NewPermissionManager(s.FgaClient)
	s.PermissionManager = permissionManager

	return nil
}

func (s *OrgPermissionService) ListenMessage() {
	subscriber := broker.NewSubscriber(s.BrokerClient.GetClient())

	subscription, appErr := subscriber.PullSubscribe(jobs.JobOrganizationRoleWildcard, constants.ConsumerOrganizationOpenFGA)
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

func (s *OrgPermissionService) processMessage(msg *nats.Msg) {
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
	case jobs.JobOrganizationMemRoleUpdated:
		processErr = s.processUpdateOrgMemRole(msg)
	case jobs.JobOrganizationMemRoleRevoked:
		processErr = s.processRevokeOrgMemRole(msg)
	case jobs.JobOrganizationMemRoleAssigned:
		processErr = s.processAssignOrgMemRole(msg)
	case jobs.JobOrganizationTeamMemRoleAssigned:
		processErr = s.processAssignTeamMemRole(msg)
	case jobs.JobOrganizationTeamMemRoleRevoked:
		processErr = s.processRevokeTeamMemRole(msg)
	case jobs.JobOrganizationTeamMemRoleUpdated:
		processErr = s.processUpdateTeamMemRole(msg)
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
