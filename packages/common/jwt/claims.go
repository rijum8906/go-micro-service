package jwt

import "time"

type Claims struct {
	UserID    string
	SessionID string
	ExpiresAt time.Time
}
