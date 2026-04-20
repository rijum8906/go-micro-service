package broker

import (
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
	stream, err := c.client.JS.StreamInfo(cfg.StreamConfig.Name)
	if err == nil {
		appErr := c.Update(cfg.StreamConfig.Name, cfg)
		if appErr != nil {
			return nil, appErr
		}
		return stream, nil
	}

	stream, err = c.client.JS.AddStream(cfg.StreamConfig)
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to create stream").
			WithDetail("stream", cfg.StreamConfig.Name).
			WithDetail("error", err.Error())
	}

	return stream, nil
}

func (c *streamManager) Update(name string, cfg *StreamConfig) *apperror.AppError {
	// TODO: implement
	return nil
}

func (c *streamManager) Delete(streamName string) *apperror.AppError {
	err := c.client.JS.DeleteStream(streamName)
	if err != nil {
		return apperror.New(apperror.CodeInternal, "Cannot delete stream "+streamName).WithDetail("error", err.Error())
	}
	return nil
}

func (c *streamManager) Get(name string) (*nats.StreamInfo, *apperror.AppError) {
	// TODO: implement
	return nil, nil
}

func (c *streamManager) Exists(name string) bool {
	// TODO: implement
	return false
}
