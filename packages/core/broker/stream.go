package broker

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
)

// StreamConfig defines JetStream stream configuration
type StreamConfig struct {
	Name       string               `validate:"required"`
	Subjects   []string             `validate:"required,min=1"`
	MaxMsgs    int64                `validate:"omitempty,gte=-1"`      // -1 = unlimited
	MaxBytes   int64                `validate:"omitempty,gte=-1"`      // -1 = unlimited
	MaxAge     time.Duration        `validate:"omitempty,gte=0"`       // 0 = unlimited
	MaxMsgSize int32                `validate:"omitempty,gte=-1"`      // -1 = unlimited
	Storage    nats.StorageType     `validate:"omitempty,oneof=0 1"`   // 0=File, 1=Memory
	Replicas   int                  `validate:"omitempty,gte=1,lte=5"` // Default 1
	Duplicates time.Duration        `validate:"omitempty,gte=0"`       // 0 = disable
	Retention  nats.RetentionPolicy `validate:"omitempty"`             // Default LimitsPolicy
	Discard    nats.DiscardPolicy   `validate:"omitempty"`             // Default DiscardOld
}

// CreateStream creates or updates a stream with persistence configuration
func (c *Client) CreateStream(cfg StreamConfig) *apperror.AppError {
	if c == nil || c.JS == nil {
		return apperror.New(apperror.CodeInternal, "jetstream client is not initialized")
	}

	err := validate.Struct(cfg)
	if err != nil {
		return apperror.New(apperror.CodeValidation, "invalid stream config").WithDetail("error", err.Error())
	}

	streamCfg := &nats.StreamConfig{
		Name:       cfg.Name,
		Subjects:   cfg.Subjects,
		MaxMsgs:    cfg.MaxMsgs,
		MaxBytes:   cfg.MaxBytes,
		MaxAge:     cfg.MaxAge,
		MaxMsgSize: cfg.MaxMsgSize,
		Storage:    cfg.Storage,
		Replicas:   cfg.Replicas,
		Duplicates: cfg.Duplicates,
		Retention:  cfg.Retention,
		Discard:    cfg.Discard,
	}

	_, err = c.JS.AddStream(streamCfg)
	if err != nil {
		return apperror.New(apperror.CodeThirdParty, "failed to create stream").
			WithDetail("stream", cfg.Name).
			WithDetail("error", err.Error())
	}

	return nil
}

func (c *Client) EnsureStream(config StreamConfig) *apperror.AppError {
	_, err := c.JS.StreamInfo(config.Name)
	if err == nil {
		return nil // Stream exists
	}

	if err != nats.ErrStreamNotFound {
		return apperror.ErrThirdParty.
			WithMessage("failed to check stream").
			WithDetail("error", err.Error())
	}

	// Stream doesn't exist, create it
	return c.CreateStream(config)
}
