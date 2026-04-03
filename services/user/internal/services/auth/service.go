package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreutils"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func (s *authService) Login(ctx context.Context, data *authv1.LoginRequest, client *dto.ClientInfo) (*authv1.AuthResponse, *apperror.AppError) {
	user, appErr := s.repos.User.GetUserByEmail(ctx, data.Email)
	if appErr != nil {
		return nil, appErr
	}

	appErr = s.utils.HashService.Verify(user.PasswordHash.String, data.Password)
	if appErr != nil {
		return nil, appErr
	}

	refreshTokenHash, appErr := s.utils.HashService.Generate(32)
	if appErr != nil {
		return nil, appErr
	}

	profile, appErr := s.repos.Profile.GetProfileByUserID(ctx, user.ID)
	if appErr != nil {
		return nil, appErr
	}

	timeStamp := pgtype.Timestamptz{
		Time:  time.Now().Add(s.env.SessionTTL),
		Valid: true,
	}
	session, appErr := s.repos.Session.CreateSession(ctx, db.CreateSessionParams{
		UserID:           user.ID,
		UserAgent:        client.UserAgent,
		IpAddr:           client.IPAddress,
		DeviceID:         client.DeviceID,
		ExpiresAt:        timeStamp,
		RefreshTokenHash: refreshTokenHash,
	})
	if appErr != nil {
		return nil, appErr
	}

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.ID.String(), session.ID.String(), token.TokenScopeAuth)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

func (s *authService) Register(ctx context.Context, data *authv1.RegisterRequest, client *dto.ClientInfo) (*authv1.AuthResponse, *apperror.AppError) {
	_, appErr := s.repos.User.GetUserByEmail(ctx, data.Email)
	if appErr != nil && appErr.Code != apperror.CodeInternal {
		return nil, appErr
	}
	if appErr == nil {
		return nil, apperror.ErrValidation.WithMessage("email already exists")
	}

	hashedPass, appErr := s.utils.HashService.Hash(data.Password)
	if appErr != nil {
		return nil, appErr
	}
	data.Password = hashedPass

	user, appErr := s.repos.User.CreateUser(ctx, &authv1.RegisterRequest{
		Email:     data.Email,
		Password:  data.Password,
		FirstName: data.FirstName,
		LastName:  data.LastName,
	})
	if appErr != nil {
		return nil, appErr
	}

	profile, appErr := s.repos.Profile.CreateProfile(ctx, db.CreateProfileParams{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		UserID:    user.ID,
	})
	if appErr != nil {
		return nil, appErr
	}

	refreshTokenHash, appErr := s.utils.HashService.Generate(32)
	if appErr != nil {
		return nil, appErr
	}

	timeStamp := pgtype.Timestamptz{
		Time:  time.Now().Add(s.env.SessionTTL),
		Valid: true,
	}
	session, appErr := s.repos.Session.CreateSession(ctx, db.CreateSessionParams{
		UserID:           user.ID,
		UserAgent:        client.UserAgent,
		IpAddr:           client.IPAddress,
		DeviceID:         client.DeviceID,
		ExpiresAt:        timeStamp,
		RefreshTokenHash: refreshTokenHash,
	})
	if appErr != nil {
		return nil, appErr
	}

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.ID.String(), session.ID.String(), token.TokenScopeAuth)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

func (s *authService) Logout(ctx context.Context, client *dto.UserInfo) (bool, *apperror.AppError) {
	sessionID, err := uuid.Parse(client.SessionID)
	if err != nil {
		return false, apperror.ErrInternal.WithMessage("Failed to parse session ID").WithDetail("error", err.Error())
	}

	appErr := s.repos.Session.RevokeSession(ctx, sessionID)
	if appErr != nil {
		return false, appErr
	}

	return true, nil
}

func (s *authService) RefreshToken(ctx context.Context, user *dto.UserInfo) (*authv1.RefreshTokenResponse, *apperror.AppError) {
	sessionID, appErr := utils.NewUUID(user.SessionID)
	if appErr != nil {
		return nil, appErr
	}

	session, appErr := s.repos.Session.GetSession(ctx, sessionID)
	if appErr != nil {
		return nil, appErr
	}

	if session.ID != sessionID {
		return nil, apperror.ErrUnAuthenticated.WithMessage("refresh token does not match session")
	}

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.UserID, session.ID.String(), token.TokenScopeAuth)
	if appErr != nil {
		return nil, appErr
	}

	return &authv1.RefreshTokenResponse{
		AccessToken: &modelsv1.Token{
			Value:     accessToken,
			ExpiresAt: coreutils.ParseToProtoTimestamp(s.env.SessionTTL),
		},
	}, nil
}

func (s *authService) RequestEmailVerification(ctx context.Context, req *authv1.RequestEmailVerificationRequest) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request email verification request is required")
	}

	user, appErr := s.repos.User.GetUserByEmail(ctx, req.GetEmail())
	if appErr != nil {
		return nil, appErr
	}

	if user.IsEmailVerified {
		return &corev1.SuccessResponse{Success: true}, nil
	}

	if _, appErr = s.utils.TokenManager.IssueScopedToken(ctx, user.ID.String(), token.TokenScopeVerifyEmail); appErr != nil {
		return nil, appErr
	}
	// TODO: Send the verification email or notification containing the scoped token.

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request password reset request is required")
	}

	user, appErr := s.repos.User.GetUserByEmail(ctx, req.GetEmail())
	if appErr != nil {
		if appErr.Code == apperror.CodeInternal {
			return &corev1.SuccessResponse{Success: true}, nil
		}
		return nil, appErr
	}

	if _, appErr = s.utils.TokenManager.IssueScopedToken(ctx, user.ID.String(), token.TokenScopeResetPassword); appErr != nil {
		return nil, appErr
	}
	// TODO: Send the password reset email or notification containing the scoped token.

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("verify email request is required")
	}

	scopedToken := req.GetScopedToken()
	if scopedToken == "" {
		return nil, apperror.ErrValidation.WithMessage("verify email scoped token is required")
	}

	claims, appErr := s.utils.TokenManager.ValidateScopedToken(ctx, scopedToken)
	if appErr != nil {
		return nil, appErr
	}

	if claims.Scope != token.TokenScopeVerifyEmail {
		return nil, apperror.ErrValidation.WithMessage("invalid scoped token scope for verify email")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	if appErr = s.repos.User.VerifyUserEmail(ctx, userID); appErr != nil {
		return nil, appErr
	}

	if appErr = s.utils.TokenManager.RevokeScopedToken(ctx, scopedToken); appErr != nil {
		return nil, appErr
	}
	// TODO: Notify the user that their email address has been verified.

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *authService) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*corev1.SuccessResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("reset password request is required")
	}

	scopedToken := req.GetScopedToken()
	if scopedToken == "" {
		return nil, apperror.ErrValidation.WithMessage("reset password scoped token is required")
	}

	claims, appErr := s.utils.TokenManager.ValidateScopedToken(ctx, scopedToken)
	if appErr != nil {
		return nil, appErr
	}

	if claims.Scope != token.TokenScopeResetPassword {
		return nil, apperror.ErrValidation.WithMessage("invalid scoped token scope for reset password")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	newPasswordHash, appErr := s.utils.HashService.Hash(req.GetNewPassword())
	if appErr != nil {
		return nil, appErr
	}

	appErr = s.repos.User.UpdateUserPassword(ctx, userID, newPasswordHash)
	if appErr != nil {
		return nil, appErr
	}

	if appErr = s.utils.TokenManager.RevokeScopedToken(ctx, scopedToken); appErr != nil {
		return nil, appErr
	}
	// TODO: Notify the user that their password has been reset.

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}
