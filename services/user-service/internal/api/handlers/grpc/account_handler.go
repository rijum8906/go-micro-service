package handlers

import (
	"context"

	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
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

	authz := md.Get(user_servicev1.XHeader_X_HEADER_USER_ID.String())
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

func (h *AuthHandler) GenerateScopedToken(ctx context.Context, req *user_servicev1.GenerateScopedTokenRequest) (*user_servicev1.GenerateScopedTokenResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	authzMetadata, err := extractAuthzMetadata(ctx)
	if err != nil {
		return nil, err
	}

	result, appErr := h.accountService.GenerateScopedToken(ctx, req, authzMetadata)
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return result, nil
}
