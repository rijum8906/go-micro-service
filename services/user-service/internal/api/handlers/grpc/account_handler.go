package handlers

import (
	"context"

	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func extractAuthzMetadata(ctx context.Context) (request.AuthzMetadata, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return request.AuthzMetadata{}, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	authz := md.Get("x-user-id")
	if len(authz) == 0 {
		return request.AuthzMetadata{}, status.Errorf(codes.Unauthenticated, "missing x-user-id")
	}

	userID, appErr := utils.StrIDToPgUUID(authz[0])
	if appErr != nil || !userID.Valid {
		return request.AuthzMetadata{}, status.Errorf(codes.Unauthenticated, "invalid user id")
	}

	return request.AuthzMetadata{
		UserID: userID,
	}, nil
}
