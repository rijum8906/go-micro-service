package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type Config struct {
	URL           string
	ClientName    string
	MaxReconnects int
	ReconnectWait time.Duration
}

type Client struct {
	Conn *nats.Conn
}

func Connect(ctx context.Context, cfg Config) (*Client, *apperror.AppError) {
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
		return nil, apperror.New(apperror.TypeThirdParty, apperror.CodeThirdParty, "failed to connect to nats").WithDetail("error", err.Error())
	}

	return &Client{
		Conn: conn,
	}, nil
}

func (c *Client) Drain() *apperror.AppError {
	if c == nil || c.Conn == nil {
		return nil
	}

	if err := c.Conn.Drain(); err != nil {
		return apperror.New(apperror.TypeInternal, apperror.CodeInternal, "failed to drain nats connection").WithDetail("error", err.Error())
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
