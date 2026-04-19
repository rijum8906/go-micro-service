package broker

import (
	"time"

	"github.com/nats-io/nats.go"
)

type StreamConfig struct {
	StreamConfig *nats.StreamConfig
}

func NewStreamConfig(name string) *StreamConfig {
	return &StreamConfig{
		StreamConfig: &nats.StreamConfig{
			Name:         name,
			MaxConsumers: 1,
			MaxMsgs:      1000,
			MaxAge:       time.Hour,
			MaxMsgSize:   1024 * 1024,
			MaxBytes:     1024 * 1024,
			Subjects:     []string{},
		},
	}
}

func (c *StreamConfig) AddDesc(desc string) *StreamConfig {
	c.StreamConfig.Description = desc
	return c
}

func (c *StreamConfig) AddSubjects(subjects ...string) *StreamConfig {
	c.StreamConfig.Subjects = subjects
	return c
}

func (c *StreamConfig) AddRetentionPolicy(policy nats.RetentionPolicy) *StreamConfig {
	c.StreamConfig.Retention = policy
	return c
}

func (c *StreamConfig) AddDiscardPolicy(policy nats.DiscardPolicy) *StreamConfig {
	c.StreamConfig.Discard = policy
	return c
}

func (c *StreamConfig) AddStorageType(storageType nats.StorageType) *StreamConfig {
	c.StreamConfig.Storage = storageType
	return c
}

func (c *StreamConfig) AddMaxConsumer(maxConsumer int) *StreamConfig {
	c.StreamConfig.MaxConsumers = maxConsumer
	return c
}

func (c *StreamConfig) AddMaxAge(age time.Duration) *StreamConfig {
	c.StreamConfig.MaxAge = age
	return c
}

func (c *StreamConfig) AddMaxMsgs(maxMsgs int64) *StreamConfig {
	c.StreamConfig.MaxMsgs = maxMsgs
	return c
}

func (c *StreamConfig) AddMaxMsgSize(maxSize int32) *StreamConfig {
	c.StreamConfig.MaxMsgSize = maxSize
	return c
}

type ConsumerConfig struct {
	ConsumerConfig *nats.ConsumerConfig
}

func NewConsumerConfig(consumerName string) *ConsumerConfig {
	return &ConsumerConfig{
		ConsumerConfig: &nats.ConsumerConfig{
			Name:          consumerName,
			Durable:       consumerName,
			DeliverPolicy: nats.DeliverAllPolicy,
			AckPolicy:     nats.AckExplicitPolicy,
			ReplayPolicy:  nats.ReplayOriginalPolicy,
		},
	}
}

func (c *ConsumerConfig) AddDeliverPolicy(policy nats.DeliverPolicy) *ConsumerConfig {
	c.ConsumerConfig.DeliverPolicy = policy
	return c
}

func (c *ConsumerConfig) AddAckPolicy(policy nats.AckPolicy) *ConsumerConfig {
	c.ConsumerConfig.AckPolicy = policy
	return c
}

func (c *ConsumerConfig) AddReplayPolicy(policy nats.ReplayPolicy) *ConsumerConfig {
	c.ConsumerConfig.ReplayPolicy = policy
	return c
}
