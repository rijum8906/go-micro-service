package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
)

func (c *Client) PullSubscribe(subject dto.JobSubject, durable, stream string) (*nats.Subscription, *apperror.AppError) {
	appErr := validateClient(c)
	if appErr != nil {
		return nil, appErr
	}

	sub, err := c.JS.PullSubscribe(string(subject), durable, nats.BindStream(stream), nats.ManualAck())
	if err != nil {
		apperror.ErrThirdParty.WithDetail("error", err.Error())
	}

	return sub, nil
}

func (c *Client) Subscribe(subject dto.JobSubject, handler func([]byte)) (*nats.Subscription, *apperror.AppError) {
	appErr := validateClient(c)
	if appErr != nil {
		return nil, appErr
	}

	if handler == nil {
		return nil, apperror.New(apperror.CodeValidation, "subscriber handler is required")
	}

	if !dto.IsValidJobSubject(string(subject)) {
		return nil, apperror.New(apperror.CodeValidation, "invalid job subject").WithDetail("subject", string(subject))
	}

	// Use JetStream subscription with manual acknowledgment
	sub, err := c.JS.Subscribe(string(subject), func(msg *nats.Msg) {
		handler(msg.Data)
		msg.Ack() // Acknowledge message after processing
	}, nats.ManualAck())
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to subscribe to nats subject").WithDetail("error", err.Error())
	}

	return sub, nil
}

// SubscribeWithOpts allows custom JetStream options
func (c *Client) SubscribeWithOpts(subject dto.JobSubject, handler func(*nats.Msg), opts ...nats.SubOpt) (*nats.Subscription, *apperror.AppError) {
	appErr := validateClient(c)
	if appErr != nil {
		return nil, appErr
	}

	if handler == nil {
		return nil, apperror.New(apperror.CodeValidation, "subscriber handler is required")
	}

	if !dto.IsValidJobSubject(string(subject)) {
		return nil, apperror.New(apperror.CodeValidation, "invalid job subject").WithDetail("subject", string(subject))
	}

	// Default to manual ack if not specified
	defaultOpts := []nats.SubOpt{nats.ManualAck()}
	defaultOpts = append(defaultOpts, opts...)

	sub, err := c.JS.Subscribe(string(subject), handler, defaultOpts...)
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to subscribe to nats subject").WithDetail("error", err.Error())
	}

	return sub, nil
}

// SubscribeDurable creates a durable subscription (survives client restarts)
func (c *Client) SubscribeDurable(subject dto.JobSubject, durableName string, handler func([]byte)) (*nats.Subscription, *apperror.AppError) {
	appErr := validateClient(c)
	if appErr != nil {
		return nil, appErr
	}

	if handler == nil {
		return nil, apperror.New(apperror.CodeValidation, "subscriber handler is required")
	}

	if !dto.IsValidJobSubject(string(subject)) {
		return nil, apperror.New(apperror.CodeValidation, "invalid job subject").WithDetail("subject", string(subject))
	}

	if durableName == "" {
		return nil, apperror.New(apperror.CodeValidation, "durable name is required")
	}

	sub, err := c.JS.Subscribe(string(subject), func(msg *nats.Msg) {
		handler(msg.Data)
		msg.Ack()
	}, nats.Durable(durableName), nats.ManualAck())
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to create durable subscription").WithDetail("error", err.Error())
	}

	return sub, nil
}

func validateClient(c *Client) *apperror.AppError {
	if c == nil || c.Conn == nil {
		return apperror.New(apperror.CodeInternal, "nats client is not initialized")
	}

	if !c.IsConnected() {
		return apperror.New(apperror.CodeThirdParty, "nats connection is not ready")
	}

	return nil
}
