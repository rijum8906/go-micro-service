package broker

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
)

type subscriber struct {
	client *brokerClient
}

func NewSubscriber(client *brokerClient) Subscriber {
	return &subscriber{
		client: client,
	}
}

// Methods

func (s *subscriber) PullSubscribe(subject dto.JobSubject, consumerName string) (*nats.Subscription, *apperror.AppError) {
	if s.client == nil {
		return nil, apperror.New(apperror.CodeValidation, "nats client is not set")
	}

	subscription, err := s.client.JS.PullSubscribe(string(subject), consumerName)
	if err != nil {
		return nil, apperror.ErrThirdParty.WithDetail("error", err.Error())
	}

	return subscription, nil
}
