package broker

import (
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type brokerClient struct {
	Conn *nats.Conn
	JS   nats.JetStreamContext
	MU   sync.RWMutex
	Addr string
}

// NewClient creates a new Message Broker Client
func NewClient() Client {
	return &brokerClient{}
}

// Connection Management

func (c *brokerClient) Connect(addr string) *apperror.AppError {
	conn, err := nats.Connect(addr)
	if err != nil {
		return apperror.ErrThirdParty.WithDetail("error", err.Error())
	}
	c.Addr = addr
	c.Conn = conn

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return apperror.New(apperror.CodeThirdParty, "failed to initialize jetstream").WithDetail("error", err.Error())
	}
	c.JS = js

	return nil
}

func (c *brokerClient) Close() *apperror.AppError {
	if c.Conn == nil {
		return apperror.New(apperror.CodeInternal, "NATS connection is not initialized yet")
	}

	c.Conn.Close()
	return nil
}

func (c *brokerClient) Drain() *apperror.AppError {
	if c.Conn == nil {
		return apperror.New(apperror.CodeInternal, "NATS connection is not initialized yet")
	}

	return nil
}

func (c *brokerClient) IsConnected() bool {
	if c.Conn == nil {
		return false
	}

	return c.Conn.IsConnected()
}

func (c *brokerClient) GetClient() *brokerClient {
	return c
}
