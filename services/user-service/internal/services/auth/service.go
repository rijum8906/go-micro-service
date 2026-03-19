package auth

import (
	"context"

	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
)

func (s *authService) Signin(ctx context.Context, req *authv1.SigninRequest) (*user_servicev1.AuthenticationResult, *errors.AppError) {
	return nil, nil
}

func (s *authService) Signup(ctx context.Context, req *authv1.SignupRequest) (*user_servicev1.AuthenticationResult, *errors.AppError) {
	return nil, nil
}

func (s *authService) Logout(ctx context.Context, req *authv1.LogoutRequest, auth request.AuthzMetadata) (*authv1.LogoutResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) LogoutAllDevices(ctx context.Context, req *authv1.LogoutAllDeviceRequest, auth request.AuthzMetadata) (*authv1.LogoutAllDeviceResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) RequestEmailVerification(ctx context.Context, req *authv1.RequestEmailVerificationRequest) *errors.AppError {
	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) *errors.AppError {
	return nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) *errors.AppError {
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) *errors.AppError {
	return nil
}

func (s *authService) ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest, authzMetadata request.AuthzMetadata) *errors.AppError {
	return nil
}

func (s *authService) GetSessions(ctx context.Context, req *authv1.GetSessionsRequest, authzMetadata request.AuthzMetadata) (*authv1.GetSessionsResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) RevokeSession(ctx context.Context, req *authv1.RevokeSessionRequest, authzMetadata request.AuthzMetadata) (*authv1.RevokeSessionResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) RevokeAllSessions(ctx context.Context, req *authv1.RevokeAllSessionsRequest, authzMetadata request.AuthzMetadata) (*authv1.RevokeAllSessionsResponse, *errors.AppError) {
	return nil, nil
}
