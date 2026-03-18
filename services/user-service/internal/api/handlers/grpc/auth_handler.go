// Package handlers contains the gRPC handlers for the auth service
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

func (h *AuthHandler) Signin(ctx context.Context, req *user_servicev1.SigninRequest) (*user_servicev1.AuthResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	result, err := h.authService.Signin(ctx, req)
	if err != nil {
		return nil, appErrorToGRPC(err)
	}

	return result, nil
}

func (h *AuthHandler) Signup(ctx context.Context, req *user_servicev1.SignupRequest) (*user_servicev1.AuthResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	result, err := h.authService.SignUp(ctx, req)
	if err != nil {
		return nil, appErrorToGRPC(err)
	}

	return result, nil
}

func (h *AuthHandler) Signout(ctx context.Context, req *user_servicev1.SignoutRequest) (*user_servicev1.SignoutResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	authz := md.Get("x-user-id")
	if len(authz) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing x-user-id")
	}

	userID, appErr := utils.StrIDToPgUUID(authz[0])
	if appErr != nil || !userID.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid user id")
	}

	authzMeta := request.AuthzMetadata{
		UserID: userID,
	}

	appErr = h.authService.Signout(ctx, req, authzMeta)
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return &user_servicev1.SignoutResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) RequestEmailVerfication(ctx context.Context, req *user_servicev1.RequestEmailVerificationRequest) (*user_servicev1.RequestEmailVerificationResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	appErr := h.authService.RequestEmailVerification(ctx, req)
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return &user_servicev1.RequestEmailVerificationResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) RequestEmailVerification(ctx context.Context, req *user_servicev1.RequestEmailVerificationRequest) (*user_servicev1.RequestEmailVerificationResponse, error) {
	return h.RequestEmailVerfication(ctx, req)
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *user_servicev1.VerifyEmailRequest) (*user_servicev1.VerifyEmailResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	appErr := h.authService.VerifyEmail(ctx, req)
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return &user_servicev1.VerifyEmailResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *user_servicev1.RequestPasswordResetRequest) (*user_servicev1.RequestPasswordResetResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	appErr := h.authService.RequestPasswordReset(ctx, req)
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return &user_servicev1.RequestPasswordResetResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *user_servicev1.ResetPasswordRequest) (*user_servicev1.ResetPasswordResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	appErr := h.authService.ResetPassword(ctx, req)
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return &user_servicev1.ResetPasswordResponse{
		Success: true,
	}, nil
}
