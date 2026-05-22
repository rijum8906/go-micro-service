package broker

import (
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type streamManager struct {
	client *brokerClient
}

// NewStreamManager creates a Stream Manager to Manager NATS streams
func NewStreamManager(client *brokerClient) StreamManager {
	return &streamManager{
		client: client,
	}
}

// Methods

func (c *streamManager) Create(cfg *StreamConfig) (*nats.StreamInfo, *apperror.AppError) {
	stream, err := c.client.JS.AddStream(cfg.StreamConfig)
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to create stream").
			WithDetail("stream", cfg.StreamConfig.Name).
			WithDetail("error", err.Error())
	}

	return stream, nil
}

func (c *streamManager) Update(cfg *StreamConfig) (*nats.StreamInfo, *apperror.AppError) {
	streamInfo, err := c.client.JS.UpdateStream(cfg.StreamConfig)
	if err != nil {
		if errors.Is(nats.ErrStreamNotFound, err) {
			return nil, apperror.New(apperror.CodeNotFound, fmt.Sprintf("stream %s not found", cfg.StreamConfig.Name)).WithDetail("error", err.Error())
		}
		return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("stream %s not found", cfg.StreamConfig.Name)).WithDetail("error", err.Error())
	}

	return streamInfo, nil
}

func (c *streamManager) Delete(streamName string) *apperror.AppError {
	err := c.client.JS.DeleteStream(streamName)
	if err != nil {
		return apperror.New(apperror.CodeInternal, "Cannot delete stream "+streamName).WithDetail("error", err.Error())
	}
	return nil
}

func (c *streamManager) Get(name string) (*nats.StreamInfo, *apperror.AppError) {
	stream, err := c.client.JS.StreamInfo(name)
	if err != nil {
		if errors.Is(nats.ErrStreamNotFound, err) {
			return nil, apperror.New(apperror.CodeNotFound, fmt.Sprintf("stream %s not found", name)).WithDetail("error", err.Error())
		}
		return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("stream %s not found", name)).WithDetail("error", err.Error())
	}

	return stream, nil
}

func (c *streamManager) Exists(name string) (bool, *apperror.AppError) {
	stream, err := c.client.JS.StreamInfo(name)
	if err != nil {
		if errors.Is(nats.ErrStreamNotFound, err) {
			return false, nil
		}
		return false, apperror.New(apperror.CodeInternal, fmt.Sprintf("stream %s not found", name)).WithDetail("error", err.Error())
	}

	return stream != nil, nil
}
