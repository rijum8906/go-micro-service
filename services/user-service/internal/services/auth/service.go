package auth

import (
	"context"
	"database/sql"
	errorsstdlib "errors"

	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

func (s *authService) Signin(ctx context.Context, req *authv1.SigninRequest) (*user_servicev1.AuthenticationResult, *errors.AppError) {
	account, appErr := s.repo.accountRepo.GetAccountByEmail(ctx, req.Email.Value)

	if appErr != nil {
		if errorsstdlib.Is(appErr.Internal, sql.ErrNoRows) {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, appErr
	}

	err := s.utilsConfig.HashService.VerifyPassword(account.PasswordHash, req.Password.Value)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	authzMetadata := request.AuthzMetadata{
		UserID: account.ID,
	}

	profiles, appErr := s.repo.profileRepo.GetProfilesByAccountID(ctx, account.ID)
	if appErr != nil {
		return nil, appErr
	}

	parsedProfiles := []*user_servicev1.Profile{}
	for _, profile := range *profiles {
		parsedProfiles = append(parsedProfiles, utils.ParseProfile(&profile))
	}

	session, appErr := s.repo.authRepo.CreateSession(ctx, request.RequestMetadata{
		UserAgent: req.Metadata.UserAgent,
		DeviceID:  req.Metadata.DeviceId,
		IPAddr:    req.Metadata.IpAddress,
	}, authzMetadata)
	if appErr != nil {
		return nil, appErr
	}

	accessToken, appErr := s.utilsConfig.JwtService.IssueToken(ctx, account.ID.String())
	if appErr != nil {
		return nil, errors.ErrInternal.WithInternal(err)
	}

	return &user_servicev1.AuthenticationResult{
		Account:  utils.ParseAccount(&account),
		Profiles: parsedProfiles,
		Tokens: &user_servicev1.AuthTokens{
			RefreshToken: utils.NewToken(session.RefreshToken, int64(s.env.ScopedJwtExpiration.Seconds())),
			AccessToken:  utils.NewToken(accessToken, int64(s.env.JwtExpiration.Seconds())),
		},
	}, nil
}

func (s *authService) Signup(ctx context.Context, req *authv1.SignupRequest) (*user_servicev1.AuthenticationResult, *errors.AppError) {
	isEmailExists, appErr := s.repo.accountRepo.IsEmailExists(ctx, req.Email.Value)
	if appErr != nil {
		return nil, appErr
	}
	if isEmailExists {
		return nil, errors.ErrConflict.WithField("email", "email already exists")
	}

	account, appErr := s.repo.accountRepo.CreateAccount(ctx, req)
	if appErr != nil {
		return nil, appErr
	}

	_, appErr = s.repo.profileRepo.CreateProfile(ctx, account.ID, req)
	if appErr != nil {
		return nil, appErr
	}

	return s.Signin(ctx, &authv1.SigninRequest{
		Email:    req.Email,
		Password: req.Password,
	})
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
