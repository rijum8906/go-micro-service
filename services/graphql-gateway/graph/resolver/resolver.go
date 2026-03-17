package resolver

import (
	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/jwt"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/generated"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/client"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Services struct {
	JWTService jwt.Service
}

type HTTPClinets struct {
	UserServiceClient *client.Client
}

type Resolver struct {
	env         *env.Env
	services    *Services
	httpClients *HTTPClinets
}

func NewResolver(env *env.Env, services *Services, httpClients *HTTPClinets) generated.ResolverRoot {
	return &Resolver{
		env:         env,
		services:    services,
		httpClients: httpClients,
	}
}
