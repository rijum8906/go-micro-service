package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/token"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/dto"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func (s *authService) Login(ctx context.Context, data *authv1.LoginRequest, client *metadata.ClientInfo) (*authv1.AuthResponse, *apperror.AppError) {
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

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.ID.String(), session.ID.String(), client.DeviceID, token.TokenScopeAuth)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

func (s *authService) Register(ctx context.Context, data *authv1.RegisterRequest, client *metadata.ClientInfo) (*authv1.AuthResponse, *apperror.AppError) {
	_, appErr := s.repos.User.GetUserByEmail(ctx, data.Email)
	if appErr != nil && appErr.Type != apperror.TypeNotFound {
		return nil, appErr
	}

	hashedPass, appErr := s.utils.HashService.Hash(data.Password)
	if appErr != nil {
		return nil, appErr
	}
	data.Password = hashedPass

	user, appErr := s.repos.User.CreateUser(ctx, &dto.Register{
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

	accessToken, appErr := s.utils.TokenManager.IssueAuthToken(ctx, user.ID.String(), session.ID.String(), client.DeviceID, token.TokenScopeAuth)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

func (s *authService) Logout(ctx context.Context, client *metadata.UserInfo) (bool, *apperror.AppError) {
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

func (s *authService) GenerateScopedToken(ctx context.Context, req *authv1.GenerateScopedTokenRequest, user *metadata.UserInfo) (*authv1.GenerateScopedTokenResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("generate scoped token request is required")
	}

	if user == nil || user.UserID == "" {
		return nil, apperror.ErrValidation.WithMessage("user metadata is required")
	}

	scopedToken, appErr := s.utils.TokenManager.IssueScopedToken(ctx, user.UserID, token.TokenScope(req.GetScope()))
	if appErr != nil {
		return nil, appErr
	}

	return &authv1.GenerateScopedTokenResponse{
		Token: &modelsv1.Token{
			Value:     scopedToken,
			ExpiresIn: timestamppb.New(time.Now().Add(s.env.ScopedTokenTTL)),
		},
	}, nil
}

func (s *authService) ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest) (*authv1.ChangePasswordResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("change password request is required")
	}

	scopedToken := req.GetScopedToken()
	if scopedToken == nil || scopedToken.GetValue() == "" {
		return nil, apperror.ErrValidation.WithMessage("change password scoped token is required")
	}

	claims, appErr := s.utils.TokenManager.ValidateScopedToken(ctx, scopedToken.GetValue())
	if appErr != nil {
		return nil, appErr
	}

	if claims.Scope != token.TokenScopeChangePassword {
		return nil, apperror.ErrValidation.WithMessage("invalid scoped token scope for change password")
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

	if appErr = s.utils.TokenManager.RevokeScopedToken(ctx, scopedToken.GetValue()); appErr != nil {
		return nil, appErr
	}

	return &authv1.ChangePasswordResponse{
		Success: true,
	}, nil
}

func (s *authService) UpdateProfileName(ctx context.Context, req *authv1.UpdateProfileNameRequest) (*modelsv1.Profile, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update profile name request is required")
	}

	profileID, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	profile, appErr := s.repos.Profile.UpdateProfileNames(ctx, profileID, req.GetFirstName(), req.GetLastName())
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(profile), nil
}

func (s *authService) UpdateProfileAvatarUrl(ctx context.Context, req *authv1.UpdateProfileAvatarUrlRequest) (*modelsv1.Profile, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("update profile avatar request is required")
	}

	profileID, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid profile id").WithDetail("error", err.Error())
	}

	profile, appErr := s.repos.Profile.UpdateProfileAvatar(ctx, profileID, req.GetAvatarUrl())
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapProfile(profile), nil
}
func (s *authService) RequestEmailVerification(ctx context.Context, req *authv1.RequestEmailVerificationRequest) (*authv1.RequestEmailVerificationResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request email verification request is required")
	}

	user, appErr := s.repos.User.GetUserByEmail(ctx, req.GetEmail())
	if appErr != nil {
		if appErr.Type == apperror.TypeNotFound {
			return &authv1.RequestEmailVerificationResponse{Success: true}, nil
		}
		return nil, appErr
	}

	if user.IsEmailVerified {
		return &authv1.RequestEmailVerificationResponse{Success: true}, nil
	}

	if _, appErr = s.utils.TokenManager.IssueScopedToken(ctx, user.ID.String(), token.TokenScopeVerifyEmail); appErr != nil {
		return nil, appErr
	}

	return &authv1.RequestEmailVerificationResponse{Success: true}, nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) (*authv1.RequestPasswordResetResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request password reset request is required")
	}

	user, appErr := s.repos.User.GetUserByEmail(ctx, req.GetEmail())
	if appErr != nil {
		if appErr.Type == apperror.TypeNotFound {
			return &authv1.RequestPasswordResetResponse{Success: true}, nil
		}
		return nil, appErr
	}

	if _, appErr = s.utils.TokenManager.IssueScopedToken(ctx, user.ID.String(), token.TokenScopeResetPassword); appErr != nil {
		return nil, appErr
	}

	return &authv1.RequestPasswordResetResponse{Success: true}, nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("verify email request is required")
	}

	scopedToken := req.GetScopedToken()
	if scopedToken == nil || scopedToken.GetValue() == "" {
		return nil, apperror.ErrValidation.WithMessage("verify email scoped token is required")
	}

	claims, appErr := s.utils.TokenManager.ValidateScopedToken(ctx, scopedToken.GetValue())
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

	if appErr = s.utils.TokenManager.RevokeScopedToken(ctx, scopedToken.GetValue()); appErr != nil {
		return nil, appErr
	}

	return &authv1.VerifyEmailResponse{
		Success: true,
	}, nil
}

func (s *authService) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*authv1.ResetPasswordResponse, *apperror.AppError) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("reset password request is required")
	}

	scopedToken := req.GetScopedToken()
	if scopedToken == nil || scopedToken.GetValue() == "" {
		return nil, apperror.ErrValidation.WithMessage("reset password scoped token is required")
	}

	claims, appErr := s.utils.TokenManager.ValidateScopedToken(ctx, scopedToken.GetValue())
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

	if appErr = s.utils.TokenManager.RevokeScopedToken(ctx, scopedToken.GetValue()); appErr != nil {
		return nil, appErr
	}

	return &authv1.ResetPasswordResponse{
		Success: true,
	}, nil
}
