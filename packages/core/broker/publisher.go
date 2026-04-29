package broker

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
)

type publisher struct {
	client *brokerClient
}

func NewPublisher(client *brokerClient) Publisher {
	return &publisher{
		client: client,
	}
}

// Methods

// Publish blocks until the subscriber send a confirmation
func (p *publisher) Publish(subject dto.JobSubject, data any) *apperror.AppError {
	if p.client == nil {
		return apperror.New(apperror.CodeValidation, "nats client is not set")
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return apperror.New(apperror.CodeInternal, "failed to marshal nats payload").WithDetail("error", err.Error())
	}

	_, err = p.client.JS.Publish(string(subject), dataBytes)
	if err != nil {
		return apperror.New(apperror.CodeThirdParty, "failed to publish to nats subject").WithDetail("error", err.Error())
	}

	return nil
}

// PublishAsync doesn't block the code
func (p *publisher) PublishAsync(subject dto.JobSubject, data any) (nats.PubAckFuture, *apperror.AppError) {
	if p.client == nil {
		return nil, apperror.New(apperror.CodeValidation, "nats client is not set")
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to marshal nats payload").WithDetail("error", err.Error())
	}

	ack, err := p.client.JS.PublishAsync(string(subject), dataBytes)
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to publish to nats subject").WithDetail("error", err.Error())
	}

	return ack, nil
}

func (p *publisher) PublishWithHeaders(subject dto.JobSubject, data any, headers nats.Header) *apperror.AppError {
	// TODO: implement
	if p.client == nil {
		return apperror.New(apperror.CodeValidation, "nats client is not set")
	}
	return nil
}
