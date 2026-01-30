package jwt

import (
	"context"
	"time"
)

type Service interface {
	IssueToken(ctx context.Context, userID string) (string, error)
	ValidateToken(ctx context.Context, token string) (*Claims, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

type Config struct {
	Issuer     string
	Secret     string
	Expiration time.Duration
}
