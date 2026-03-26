package dto

import (
	"time"

	"github.com/rijum8906/relay/services/user/internal/db"
)

type Tokens struct {
	AccessToken  TokenResult
	RefreshToken TokenResult
}

type TokenResult struct {
	Value     string
	ExpiresAt time.Time
}

type AuthResult struct {
	User    *db.User
	Profile *db.Profile
	Tokens  *Tokens
}
