package auth

import (
	"context"

	"github.com/rijum8906/relay/packages/common/errors"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
)

func (s *authService) Signin(ctx context.Context, req *user_servicev1.SigninRequest) (*user_servicev1.AuthResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) Signup(ctx context.Context, req *user_servicev1.SignupRequest) (*user_servicev1.AuthResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) Logout(ctx context.Context, req *user_servicev1.SignoutRequest, auth request.AuthzMetadata) (*user_servicev1.SignoutResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) LogoutAllDevices(ctx context.Context, req *user_servicev1.SignoutAllRequest, auth request.AuthzMetadata) (*user_servicev1.SignoutAllResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) RequestEmailVerification(ctx context.Context, req *user_servicev1.RequestEmailVerificationRequest) *errors.AppError {
	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *user_servicev1.VerifyEmailRequest) *errors.AppError {
	return nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, req *user_servicev1.RequestPasswordResetRequest) *errors.AppError {
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req *user_servicev1.ResetPasswordRequest) *errors.AppError {
	return nil
}

func (s *authService) ChangePassword(ctx context.Context, req *user_servicev1.ChangePasswordRequest, authzMetadata request.AuthzMetadata) *errors.AppError {
	return nil
}

func (s *authService) GetSessions(ctx context.Context, req *user_servicev1.GetSessionsRequest, authzMetadata request.AuthzMetadata) (*user_servicev1.GetSessionsResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) RevokeSession(ctx context.Context, req *user_servicev1.RevokeSessionRequest, authzMetadata request.AuthzMetadata) (*user_servicev1.RevokeSessionResponse, *errors.AppError) {
	return nil, nil
}

func (s *authService) RevokeAllSessions(ctx context.Context, req *user_servicev1.RevokeAllSessionsRequest, authzMetadata request.AuthzMetadata) (*user_servicev1.RevokeAllSessionsResponse, *errors.AppError) {
	return nil, nil
}
