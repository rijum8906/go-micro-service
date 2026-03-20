package resolver

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	commonv1 "github.com/rijum8906/relay/packages/pb/common/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/model"
	"google.golang.org/grpc/metadata"
)

func requestMetadataFromContext(ctx context.Context, input *model.MetadataInput) *commonv1.RequestMetadata {
	if input == nil {
		return nil
	}

	metadata := &commonv1.RequestMetadata{
		DeviceId: input.DeviceID,
	}

	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return metadata
	}

	metadata.UserAgent = opCtx.Headers.Get("User-Agent")
	metadata.IpAddress = opCtx.Headers.Get("X-Client-IP")

	return metadata
}

func withAuthzMetadata(ctx context.Context) context.Context {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return ctx
	}

	userID := opCtx.Headers.Get("X-User-ID")
	if userID == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, "x-user-id", userID)
}
