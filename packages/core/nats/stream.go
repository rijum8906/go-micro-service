package nats

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
)

// StreamConfig defines JetStream stream configuration
type StreamConfig struct {
	Name       string
	Subjects   []string
	MaxMsgs    int64                // Max messages in stream
	MaxBytes   int64                // Max total bytes
	MaxAge     time.Duration        // How long to keep messages
	MaxMsgSize int32                // Max individual message size
	Storage    nats.StorageType     // FileStorage (recommended) or MemoryStorage
	Replicas   int                  // For clustering (1-5)
	Duplicates time.Duration        // Duplicate detection window
	Retention  nats.RetentionPolicy // Limits, Interest, or WorkQueue
	Discard    nats.DiscardPolicy   // DiscardOld or DiscardNew
}

// CreateStream creates or updates a stream with persistence configuration
func (c *Client) CreateStream(cfg StreamConfig) *apperror.AppError {
	if c == nil || c.JS == nil {
		return apperror.New(apperror.CodeInternal, "jetstream client is not initialized")
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

	// Set defaults if not provided
	if streamCfg.Storage == 0 {
		streamCfg.Storage = nats.FileStorage
	}
	if streamCfg.Replicas == 0 {
		streamCfg.Replicas = 1
	}
	if streamCfg.Retention == 0 {
		streamCfg.Retention = nats.LimitsPolicy
	}
	if streamCfg.Discard == 0 {
		streamCfg.Discard = nats.DiscardOld
	}

	_, err := c.JS.AddStream(streamCfg)
	if err != nil {
		return apperror.New(apperror.CodeThirdParty, "failed to create stream").
			WithDetail("stream", cfg.Name).
			WithDetail("error", err.Error())
	}

	return nil
}
