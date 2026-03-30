package resolver

import (
	"github.com/go-playground/validator/v10"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	AuthClient authv1.AuthServiceClient
	Validate   *validator.Validate
	Token      *token.TokenManager
}

func NewResolver(authClient authv1.AuthServiceClient, token *token.TokenManager) *Resolver {
	return &Resolver{
		Validate:   validator.New(),
		AuthClient: authClient,
		Token:      token,
	}
}
