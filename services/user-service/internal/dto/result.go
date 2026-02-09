package dto

import db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"

type MyAccountResult struct {
	Account         *db.Account         `json:"account"`
	Profiles        *[]db.Profile       `json:"profiles"`
	AccountSecurity *db.AccountSecurity `json:"accountSecurity"`
	OAuths          *[]db.Oauth         `json:"oAuths"`
}

type GetProfileResult struct {
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	DisplayName *string `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
}
