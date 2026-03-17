package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/generated"
)

const (
	authHeader  = "X-Is-Authenticated"
	rolesHeader = "X-Roles"
)

func NewDirectiveRoot() generated.DirectiveRoot {
	return generated.DirectiveRoot{
		Authenticated: authenticatedDirective,
		HasRole:       hasRoleDirective,
		Public:        publicDirective,
		RateLimit:     rateLimitDirective,
	}
}

func publicDirective(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return next(ctx)
}

func authenticatedDirective(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	if !isAuthenticated(ctx) {
		return nil, fmt.Errorf("authentication required")
	}
	return next(ctx)
}

func hasRoleDirective(ctx context.Context, obj any, next graphql.Resolver, role string) (any, error) {
	if !isAuthenticated(ctx) {
		return nil, fmt.Errorf("authentication required")
	}
	if role == "" {
		return next(ctx)
	}

	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return nil, fmt.Errorf("operation context not available")
	}

	rawRoles := opCtx.Headers.Values(rolesHeader)
	for _, raw := range rawRoles {
		for _, candidate := range strings.Split(raw, ",") {
			if strings.EqualFold(strings.TrimSpace(candidate), role) {
				return next(ctx)
			}
		}
	}

	return nil, fmt.Errorf("forbidden")
}

func rateLimitDirective(ctx context.Context, obj any, next graphql.Resolver, limit int32, duration int32) (any, error) {
	// Keep schema directives wired even before rate limiting is backed by storage.
	return next(ctx)
}

func isAuthenticated(ctx context.Context) bool {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return false
	}
	return strings.EqualFold(opCtx.Headers.Get(authHeader), "true")
}
