package resolver

import (
	"github.com/rijum8906/relay/packages/common/env"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/client"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Env        *env.Env
	UserClient user_servicev1.AuthServiceClient
}

func NewResolver(env *env.Env) *Resolver {
	userClient := client.NewUserClient()
	return &Resolver{
		Env:        env,
		UserClient: userClient,
	}
}
