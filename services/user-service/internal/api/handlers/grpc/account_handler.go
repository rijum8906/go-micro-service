package handlers

import (
	"context"

	accountv1 "github.com/rijum8906/relay/packages/pb/user_service/account/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (h *AccountHandler) GetAccount(ctx context.Context, req *accountv1.GetAccountRequest) (*accountv1.GetAccountResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	accountID, appErr := utils.StrIDToPgUUID(req.GetAccountId().GetValue())
	if appErr != nil || !accountID.Valid {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account id")
	}

	account, appErr := h.accountService.MyAccount(ctx, &request.AuthzMetadata{UserID: accountID})
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return &accountv1.GetAccountResponse{
		Account: utils.ParseAccount(account),
	}, nil
}

func (h *AccountHandler) IsAccountExistsByEmail(ctx context.Context, req *accountv1.IsAccountExistsByEmailRequest) (*accountv1.IsAccountExistsByEmailResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	exists, appErr := h.accountService.IsEmailExists(ctx, req.GetEmail().GetValue())
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return &accountv1.IsAccountExistsByEmailResponse{
		Exists: exists,
	}, nil
}

func (h *AccountHandler) MyAccount(ctx context.Context, req *accountv1.MyAccountRequest) (*accountv1.MyAccountResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	authzMetadata, err := extractAuthzMetadata(ctx)
	if err != nil {
		return nil, err
	}

	account, appErr := h.accountService.MyAccount(ctx, &authzMetadata)
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return &accountv1.MyAccountResponse{
		Account: utils.ParseAccount(account),
	}, nil
}

func (h *AccountHandler) UpdateEmail(ctx context.Context, req *accountv1.UpdateEmailRequest) (*accountv1.UpdateEmailResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateEmail not implemented")
}

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
