package dto

import db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type AuthenticationResult struct {
	Account  *db.Account
	Profiles []*db.Profile
	Tokens   *Tokens
}
