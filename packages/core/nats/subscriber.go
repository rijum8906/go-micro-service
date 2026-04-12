package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
)

func (c *Client) Subscribe(subject string, handler func([]byte)) (*nats.Subscription, *apperror.AppError) {
	if c == nil || c.Conn == nil {
		return nil, apperror.New(apperror.CodeInternal, "nats client is not initialized")
	}

	if !c.IsConnected() {
		return nil, apperror.New(apperror.CodeThirdParty, "nats connection is not ready")
	}

	if handler == nil {
		return nil, apperror.New(apperror.CodeValidation, "subscriber handler is required")
	}

	if !dto.IsValidJobSubject(subject) {
		return nil, apperror.New(apperror.CodeValidation, "invalid job subject").WithDetail("subject", subject)
	}

	sub, err := c.Conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to subscribe to nats subject").WithDetail("error", err.Error())
	}

	return sub, nil
}
