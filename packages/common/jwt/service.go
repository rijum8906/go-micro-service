package jwt

import "context"

type Service interface {
	IssueToken(ctx context.Context, userID string) (string, error)
	ValidateToken(ctx context.Context, token string) (*Claims, error)
	RevokeSession(ctx context.Context, sessionID string) error
}
