// Package broker
package broker

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type Client interface {
	Connect(addr string) *apperror.AppError
	Drain() *apperror.AppError
	IsConnected() bool
	Close() *apperror.AppError
	GetClient() *brokerClient
}

type Publisher interface {
	Publish(subject string, data any) *apperror.AppError
	PublishAsync(subject string, data any) (nats.PubAckFuture, *apperror.AppError)
	PublishWithHeaders(subject string, data any, headers nats.Header) *apperror.AppError
}

type Subscriber interface {
	PullSubscribe(subject string, consumerName string) (*nats.Subscription, *apperror.AppError)
}

type StreamManager interface {
	Create(config *StreamConfig) (*nats.StreamInfo, *apperror.AppError)
	Update(streamName string, config *StreamConfig) *apperror.AppError
	Delete(streamName string) *apperror.AppError
	Get(streamName string) (*nats.StreamInfo, *apperror.AppError)
	Exists(streamName string) bool
}

type ConsumerManager interface {
	Create(streamName string, config *ConsumerConfig) (*nats.ConsumerInfo, *apperror.AppError)
	Delete(streamName, consumerName string) *apperror.AppError
	Get(streamName, consumerName string) (*nats.ConsumerInfo, *apperror.AppError)
}
