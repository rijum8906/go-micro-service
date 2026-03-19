// Package handlers contains the gRPC handlers for the auth service
package handlers

import (
	"context"
	"errors"

	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *AuthHandler) Signin(ctx context.Context, req *authv1.SigninRequest) (*authv1.SigninResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	result, err := h.authService.Signin(ctx, req)
	if err != nil {
		return nil, appErrorToGRPC(err)
	}

	return &authv1.SigninResponse{
		Result: result,
	}, nil
}

func (h *AuthHandler) Signup(ctx context.Context, req *authv1.SignupRequest) (*authv1.SignupResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	result, err := h.authService.Signup(ctx, req)
	if err != nil {
		return nil, appErrorToGRPC(err)
	}

	return &authv1.SignupResponse{
		Result: result,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	authzMetadata, err := extractAuthzMetadata(ctx)
	if err != nil {
		return nil, err
	}

	result, appErr := h.authService.Logout(ctx, req, authzMetadata)
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return result, nil
}

func (h *AuthHandler) LogoutAllDevice(ctx context.Context, req *authv1.LogoutAllDeviceRequest) (*authv1.LogoutAllDeviceResponse, error) {
	if err := h.validator.Validate(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	authzMetadata, err := extractAuthzMetadata(ctx)
	if err != nil {
		return nil, err
	}

	result, appErr := h.authService.LogoutAllDevices(ctx, req, authzMetadata)
	if appErr != nil {
		return nil, appErrorToGRPC(appErr)
	}

	return result, nil
}

func (h *AuthHandler) RequestEmailVerification(ctx context.Context, req *authv1.RequestEmailVerificationRequest) (*authv1.RequestEmailVerificationResponse, error) {
	return nil, errors.New("fucking error")
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	return nil, errors.New("fucking error")
}

func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) (*authv1.RequestPasswordResetResponse, error) {
	return nil, errors.New("fucking error")
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*authv1.ResetPasswordResponse, error) {
	return nil, errors.New("fucking error")
}

func (h *AuthHandler) ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest) (*authv1.ChangePasswordResponse, error) {
	return nil, errors.New("fucking error")
}

func (h *AuthHandler) GetSessions(ctx context.Context, req *authv1.GetSessionsRequest) (*authv1.GetSessionsResponse, error) {
	return nil, errors.New("fucking error")
}

func (h *AuthHandler) RevokeSession(ctx context.Context, req *authv1.RevokeSessionRequest) (*authv1.RevokeSessionResponse, error) {
	return nil, errors.New("fucking error")
}

func (h *AuthHandler) RevokeAllSessions(ctx context.Context, req *authv1.RevokeAllSessionsRequest) (*authv1.RevokeAllSessionsResponse, error) {
	return nil, errors.New("fucking error")
}

func (h *AuthHandler) GenerateScopedToken(ctx context.Context, req *authv1.GenerateScopedTokenRequest) (*authv1.GenerateScopedTokenResponse, error) {
	return nil, errors.New("fucking error")
}
