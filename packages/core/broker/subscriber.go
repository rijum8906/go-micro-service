package broker

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

func validateClient(c *Client) *apperror.AppError {
	if c == nil || c.Conn == nil {
		return apperror.New(apperror.CodeInternal, "nats client is not initialized")
	}

	if !c.IsConnected() {
		return apperror.New(apperror.CodeThirdParty, "nats connection is not ready")
	}

	return nil
}
