package resolver

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/generated"
)

func NewDirectiveRoot() generated.DirectiveRoot {
	return generated.DirectiveRoot{
		Authenticated: authenticatedDirective,
		HasRole:       passthroughRoleDirective,
		Public:        passthroughDirective,
		RateLimit:     passthroughRateLimitDirective,
	}
}

func authenticatedDirective(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil || opCtx.Headers.Get("X-Is-Authenticated") != "true" {
		return nil, errors.New("authentication required")
	}

	return next(ctx)
}

func passthroughDirective(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	return next(ctx)
}

func passthroughRateLimitDirective(ctx context.Context, _ any, next graphql.Resolver, _ int32, _ int32) (any, error) {
	return next(ctx)
}

func passthroughRoleDirective(ctx context.Context, _ any, next graphql.Resolver, _ string) (any, error) {
	return next(ctx)
}
