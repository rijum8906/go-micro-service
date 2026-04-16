package nats

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
)

func (c *Client) Publish(subject dto.JobSubject, payload []byte) *apperror.AppError {
	if c == nil || c.Conn == nil {
		return apperror.New(apperror.CodeInternal, "nats client is not initialized")
	}

	if !c.IsConnected() {
		return apperror.New(apperror.CodeThirdParty, "nats connection is not ready")
	}

	if !dto.IsValidJobSubject(string(subject)) {
		return apperror.New(apperror.CodeValidation, "invalid job subject").WithDetail("subject", string(subject))
	}

	// Use JetStream for guaranteed delivery
	if _, err := c.JS.Publish(string(subject), payload); err != nil {
		return apperror.New(apperror.CodeThirdParty, "failed to publish nats message").WithDetail("error", err.Error())
	}

	return nil
}

func (c *Client) PublishJSON(subject dto.JobSubject, payload any) *apperror.AppError {
	raw, err := json.Marshal(payload)
	if err != nil {
		return apperror.New(apperror.CodeInternal, "failed to marshal nats payload").WithDetail("error", err.Error())
	}

	return c.Publish(subject, raw)
}

func (c *Client) PublishWithAck(subject dto.JobSubject, payload []byte) (*nats.PubAck, *apperror.AppError) {
	if c == nil || c.Conn == nil {
		return nil, apperror.New(apperror.CodeInternal, "nats client is not initialized")
	}

	if !c.IsConnected() {
		return nil, apperror.New(apperror.CodeThirdParty, "nats connection is not ready")
	}

	if !dto.IsValidJobSubject(string(subject)) {
		return nil, apperror.New(apperror.CodeValidation, "invalid job subject").WithDetail("subject", string(subject))
	}

	ack, err := c.JS.Publish(string(subject), payload)
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to publish nats message").WithDetail("error", err.Error())
	}

	return ack, nil
}
