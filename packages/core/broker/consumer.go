package broker

import (
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

func (m *consumerManager) Create(steamName string, config *ConsumerConfig) (*nats.ConsumerInfo, *apperror.AppError) {
	consumer, err := m.client.JS.AddConsumer(steamName, config.ConsumerConfig)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to create consumer").WithDetail("error", err.Error())
	}

	return consumer, nil
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
		return nil, apperror.New(apperror.CodeInternal, "failed to get consumer").WithDetail("error", err.Error())
	}

	return consumer, nil
}
