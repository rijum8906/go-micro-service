package broker

import (
	"errors"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type consumerManager struct {
	client *brokerClient
	mu     sync.RWMutex
}

func NewConsumerManager(client *brokerClient) ConsumerManager {
	return &consumerManager{
		client: client,
	}
}

func (m *consumerManager) Create(steamName string, cfg *ConsumerConfig) (*nats.ConsumerInfo, *apperror.AppError) {
	consumer, err := m.client.JS.AddConsumer(steamName, cfg.ConsumerConfig)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to create consumer").WithDetail("error", err.Error())
	}

	return consumer, nil
}

func (m *consumerManager) Update(streamName string, cfg *ConsumerConfig) (*nats.ConsumerInfo, *apperror.AppError) {
	consumerOnfo, err := m.client.JS.UpdateConsumer(streamName, cfg.ConsumerConfig)
	if err != nil {
		if errors.Is(nats.ErrConsumerNotFound, err) {
			return nil, apperror.New(apperror.CodeNotFound, fmt.Sprintf("consumer %s not found in stream %s", cfg.ConsumerConfig.Name, streamName)).WithDetail("error", err.Error())
		}
		return nil, apperror.New(apperror.CodeInternal, "failed to get consumer").WithDetail("error", err.Error())
	}

	return consumerOnfo, nil
}

func (m *consumerManager) Delete(streamName, consumerName string) *apperror.AppError {
	err := m.client.JS.DeleteConsumer(streamName, consumerName)
	if err != nil {
		return apperror.New(apperror.CodeInternal, "failed to delete consumer").WithDetail("error", err.Error())
	}
	return nil
}

func (m *consumerManager) Get(streamName, consumerName string) (*nats.ConsumerInfo, *apperror.AppError) {
	consumer, err := m.client.JS.ConsumerInfo(streamName, consumerName)
	if err != nil {
		if errors.Is(nats.ErrConsumerNotFound, err) {
			return nil, apperror.New(apperror.CodeNotFound, fmt.Sprintf("consumer %s not found in stream %s", consumerName, streamName)).WithDetail("error", err.Error())
		}
		return nil, apperror.New(apperror.CodeInternal, "failed to get consumer").WithDetail("error", err.Error())
	}

	return consumer, nil
}

func (m *consumerManager) Exists(streamName, consumerName string) (bool, *apperror.AppError) {
	info, err := m.client.JS.ConsumerInfo(streamName, consumerName)
	if err != nil {
		if errors.Is(nats.ErrConsumerNotFound, err) {
			return false, nil
		}
		return false, apperror.New(apperror.CodeInternal, "failed to get consumer").WithDetail("error", err.Error())
	}

	return info != nil, nil
}
