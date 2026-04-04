package nats

import (
	"encoding/json"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
)

func (c *Client) Publish(subject string, payload []byte) *apperror.AppError {
	if c == nil || c.Conn == nil {
		return apperror.New(apperror.CodeInternal, "nats client is not initialized")
	}

	if !c.IsConnected() {
		return apperror.New(apperror.CodeThirdParty, "nats connection is not ready")
	}

	if !dto.IsValidJobSubject(subject) {
		return apperror.New(apperror.CodeValidation, "invalid job subject").WithDetail("subject", subject)
	}

	if err := c.Conn.Publish(subject, payload); err != nil {
		return apperror.New(apperror.CodeThirdParty, "failed to publish nats message").WithDetail("error", err.Error())
	}

	return nil
}

func (c *Client) PublishJSON(subject string, payload any) *apperror.AppError {
	raw, err := json.Marshal(payload)
	if err != nil {
		return apperror.New(apperror.CodeInternal, "failed to marshal nats payload").WithDetail("error", err.Error())
	}

	return c.Publish(subject, raw)
}

func (c *Client) PublishEmail(message dto.EmailMetadata) *apperror.AppError {
	if appErr := message.Validate(); appErr != nil {
		return appErr
	}

	return c.PublishJSON(message.JobSubject.String(), message)
}

func (c *Client) PublishNotification(message dto.NotificationMessage) *apperror.AppError {
	if appErr := message.Validate(); appErr != nil {
		return appErr
	}

	return c.PublishJSON(message.JobSubject.String(), message)
}
