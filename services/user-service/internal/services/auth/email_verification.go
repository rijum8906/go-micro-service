package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/jobs"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/utils"
	"go.uber.org/zap"
)

// NOTE: no need to validate any request parameters only validate if the request is not nil
// validation will be handled by gateway

// RequestEmailVerification handles the request for email verification
func (s *AuthService) RequestEmailVerification(ctx context.Context, req *authv1.RequestEmailVerificationRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request email verification request is required")
	}

	var (
		verificationURL string
		profile         *db.Profile
		user            *db.User
	)

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Get user by email
		u, err := q.GetUserByEmail(ctx, req.Email)
		if appErr := utils.AssertRowExists(err, "user", req.Email); appErr != nil {
			return appErr
		}
		user = &u

		// Idempotent check: if email is already verified, return success
		if u.IsEmailVerified {
			return nil
		}

		// Get profile by user ID
		p, err := q.GetProfileByUserID(ctx, user.ID)
		if appErr := utils.AssertRowExists(err, "profile", user.ID.String()); appErr != nil {
			return appErr
		}
		profile = &p

		// Issue scoped token for email verification
		scopedToken, appErr := s.TokenManager.IssueScopedToken(ctx, user.ID.String(), constants.TokenScopeVerifyEmail)
		if appErr != nil {
			return appErr
		}

		// Generate verification URL using the scoped token
		url, appErr := utils.NewTokenURL(scopedToken.TokenString, s.Config.FrontendURL, s.Config.EmailVerificationPath)
		if appErr != nil {
			return appErr
		}
		verificationURL = url

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	if appErr := s.BrokerPublisher.Publish(jobs.JobUserRequestedEmailVerification, dto.EmailVerificationDTO{
		BaseEmailDTO: dto.BaseEmailDTO{
			ClientName:  profile.FirstName + " " + profile.LastName,
			ClientEmail: user.Email,
		},
		VerificationURL: verificationURL,
		Validity:        "10 minutes",
	}); appErr != nil {
		s.Logger.Error("failed to publish email verification job", zap.Error(appErr),
			zap.String("email", user.Email), zap.String("client_name", profile.FirstName+" "+profile.LastName))
		return nil, appErr
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("request password reset request is required")
	}

	var (
		resetURL string
		profile  *db.Profile
		user     *db.User
	)

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		u, err := q.GetUserByEmail(ctx, req.Email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrNotFound.WithMessage("user not found")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to get user by email").
				WithDetail("db_error", err.Error())
		}
		if u.IsEmailVerified {
			return nil
		}

		user = &u

		p, err := q.GetProfileByUserID(ctx, user.ID)
		if appErr := utils.AssertRowExists(err, "profile", user.ID.String()); appErr != nil {
			return appErr
		}
		profile = &p

		scopedToken, appErr := s.TokenManager.IssueScopedToken(ctx, user.ID.String(), constants.TokenScopeVerifyEmail)
		if appErr != nil {
			return appErr
		}

		url, appErr := utils.NewTokenURL(scopedToken.TokenString, s.Config.FrontendURL, s.Config.ResetPasswordPath)
		if appErr != nil {
			return appErr
		}
		resetURL = url

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	if appErr := s.BrokerPublisher.Publish(jobs.JobUserRequestedPasswordReset, dto.PasswordResetDTO{
		BaseEmailDTO: dto.BaseEmailDTO{
			ClientName:  profile.FirstName + " " + profile.LastName,
			ClientEmail: user.Email,
		},
		ResetURL: resetURL,
		Validity: "10 minutes",
	}); appErr != nil {
		s.Logger.Error("failed to publish password reset job", zap.Error(appErr),
			zap.String("email", user.Email), zap.String("client_name", profile.FirstName+" "+profile.LastName))
		return nil, appErr
	}

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("verify email request is required")
	}

	claims, appErr := s.TokenManager.ValidateScopedToken(ctx, req.ScopedToken)
	if appErr != nil {
		return nil, appErr
	}

	if claims.Scope != constants.TokenScopeVerifyEmail {
		return nil, apperror.ErrValidation.WithMessage("invalid scoped token scope for verify email")
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id in token")
	}

	if appErr = s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		user, err := q.GetUser(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrNotFound.WithMessage("user not found")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to get user by id").
				WithDetail("db_error", err.Error())
		}

		if user.IsEmailVerified {
			return nil
		}

		if err = q.UpdateUserIsEmailVerified(ctx, db.UpdateUserIsEmailVerifiedParams{
			IsEmailVerified: true,
			ID:              userID,
		}); err != nil {
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to update user").
				WithDetail("db_error", err.Error())
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// TODO: Notify the user that their email address has been verified.

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*corev1.SuccessResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("reset password request is required")
	}

	claims, appErr := s.TokenManager.ValidateScopedToken(ctx, req.ScopedToken)
	if appErr != nil {
		return nil, appErr
	}

	if claims.Scope != constants.TokenScopeResetPassword {
		return nil, apperror.ErrValidation.WithMessage("invalid scoped token scope for reset password")
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id in token")
	}

	if appErr = s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		user, err := q.GetUser(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrNotFound.WithMessage("user not found")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to get user by id").
				WithDetail("db_error", err.Error())
		}

		if user.IsEmailVerified {
			return nil
		}

		passwdHash, appErr := s.HashService.Hash(req.NewPassword)
		if appErr != nil {
			return appErr
		}

		if err = q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
			ID: userID,
			PasswordHash: pgtype.Text{
				String: passwdHash,
				Valid:  true,
			},
		}); err != nil {
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to update user").
				WithDetail("db_error", err.Error())
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// TODO: Notify the user that their password has been reset.

	return &corev1.SuccessResponse{
		Success: true,
	}, nil
}
