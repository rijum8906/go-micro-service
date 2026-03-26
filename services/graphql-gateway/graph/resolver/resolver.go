package resolver

import (
	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	AuthClient authv1.AuthServiceClient
}

func NewResolver(authClient authv1.AuthServiceClient) *Resolver {
	return &Resolver{
		AuthClient: authClient,
	}
}
