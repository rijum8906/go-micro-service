// Package broker
package broker

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
)

var validate = validator.New()

type Config struct {
	URL           string        `validate:"required,url"`
	ClientName    string        `validate:"required"`
	MaxReconnects int           `validate:"gte=0"`
	ReconnectWait time.Duration `validate:"gte=0"`
}

type Client struct {
	Conn *nats.Conn
	JS   nats.JetStreamContext
}

func Connect(ctx context.Context, cfg Config) (*Client, *apperror.AppError) {
	if err := validate.Struct(cfg); err != nil {
		return nil, apperror.ErrValidation.WithDetail("error", err.Error())
	}

	if err := ctx.Err(); err != nil {
		return nil, apperror.ErrInternal.WithMessage("nats connection cancelled by context")
	}

	url := cfg.URL
	if url == "" {
		url = nats.DefaultURL
	}

	opts := []nats.Option{}

	if cfg.ClientName != "" {
		opts = append(opts, nats.Name(cfg.ClientName))
	}

	if cfg.MaxReconnects != 0 {
		opts = append(opts, nats.MaxReconnects(cfg.MaxReconnects))
	}

	if cfg.ReconnectWait > 0 {
		opts = append(opts, nats.ReconnectWait(cfg.ReconnectWait))
	}

	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, apperror.New(apperror.CodeThirdParty, "failed to connect to nats").WithDetail("error", err.Error())
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, apperror.New(apperror.CodeThirdParty, "failed to initialize jetstream").WithDetail("error", err.Error())
	}

	return &Client{
		Conn: conn,
		JS:   js,
	}, nil
}

func (c *Client) Drain() *apperror.AppError {
	if c == nil || c.Conn == nil {
		return nil
	}

	// This blocks until drain is complete
	if err := c.Conn.Drain(); err != nil {
		return apperror.New(apperror.CodeInternal, "failed to drain nats connection").WithDetail("error", err.Error())
	}

	return nil
}

func (c *Client) Close() {
	if c == nil || c.Conn == nil {
		return
	}

	c.Conn.Close()
}

func (c *Client) IsConnected() bool {
	return c != nil && c.Conn != nil && c.Conn.IsConnected()
}

func (c *Client) JetStream() nats.JetStreamContext {
	return c.JS
}

func (c *Client) Connection() *nats.Conn {
	return c.Conn
}
