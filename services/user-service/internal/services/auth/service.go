package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreutils"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/jobs"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/utils"
	"go.uber.org/zap"
)

// NOTE: no need to validate any request parameters only validate if the request is not nil
// validation will be handled by gateway

// Login authenticates a user with email and password, creating a new session.
//
// This method handles user authentication by validating credentials, creating a session,
// and issuing access and refresh tokens for subsequent authenticated requests.
//
// Security Constraints:
//   - Password verification is constant-time to prevent timing attacks
//   - Invalid credentials return generic error (prevents user enumeration)
//   - Refresh tokens are hashed before storage (never stored in plaintext)
//   - Access tokens are short-lived (configurable TTL)
//
// Error Responses:
//   - Validation: Invalid request or missing fields
//   - PermissionDenied: Invalid credentials (generic error)
//   - Internal: Database error, token generation failure, or missing profile
//
// Example:
//
//	resp, err := service.Login(ctx, &authv1.LoginRequest{
//	    Email:    "user@example.com",
//	    Password: "securePassword123!",
//	})
func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	// Validate request
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("login request is required")
	}

	// Extract client information for device fingerprinting
	clientInfo, ok := metadata.ReceiveClientInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "client info not found in context")
	}

	var (
		user             *db.User
		profile          *db.Profile
		refreshTokenHash string
		accessToken      string
	)

	// Execute authentication in transaction
	if err := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Retrieve user by email
		u, err := q.GetUserByEmail(ctx, req.GetEmail())
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Generic error prevents user enumeration
				return apperror.ErrPermissionDenied.WithMessage("invalid credentials")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to get user by email").
				WithDetail("db_error", err.Error())
		}
		user = &u

		// OAuth users without password cannot use password login
		if !user.PasswordHash.Valid {
			return apperror.ErrPermissionDenied.WithMessage("invalid credentials")
		}

		// Verify password
		if !s.HashService.Verify(user.PasswordHash.String, req.GetPassword()) {
			return apperror.ErrPermissionDenied.WithMessage("invalid credentials")
		}

		// Retrieve user profile
		p, err := q.GetProfileByUserID(ctx, user.ID)
		if err != nil {
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to get user profile").
				WithDetail("db_error", err.Error())
		}
		profile = &p

		// Create session
		session, appErr := s.createSession(ctx, q, user.ID, refreshTokenHash, &clientInfo)
		if appErr != nil {
			return appErr
		}

		// Issue access token and refresh token
		tokenPair, appErr := s.issueTokenPair(ctx, user.ID.String(), session.ID.String())
		if appErr != nil {
			return appErr
		}
		accessToken = tokenPair.AccessToken
		refreshTokenHash = tokenPair.RefreshTokenHash

		return nil
	}); err != nil {
		return nil, err
	}

	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

// Register creates a new user account and initiates the authentication flow.
//
// This method handles new user registration by validating input, creating user records,
// establishing a session, issuing tokens, and triggering email verification.
//
// Execution Flow:
//   - Validate request parameters (email, password, name fields)
//   - Check if email already exists (prevent duplicate registration)
//   - Hash the user's password using bcrypt
//   - Begin database transaction
//   - Create user record in database
//   - Create user profile record
//   - Generate refresh token hash
//   - Create session record with device information
//   - Issue access token (JWT)
//   - Commit transaction
//   - Trigger email verification (asynchronous)
//   - Return authentication response with tokens and user info
//
// Security Constraints:
//   - Email must be unique across all users
//   - Password must meet complexity requirements (validated by HashService)
//   - Passwords are never stored in plaintext
//   - Refresh tokens are hashed before storage
//   - Access tokens are short-lived JWTs
//   - Email verification required before certain operations
//
// Email Uniqueness:
//   - Check is performed before any database writes
//   - Race condition possible between check and insert
//   - Database unique constraint provides final protection
//   - Returns Conflict error if email already exists
//
// Transaction Boundaries:
//   - User creation, profile creation, and session creation in atomic transaction
//   - Ensures all or none of the operations succeed
//   - Prevents orphaned records if any step fails
//
// Email Verification:
//   - Triggered asynchronously (non-blocking)
//   - User can log in immediately but may have restricted access
//   - Verification email sent to user's email address
//   - Verification token expires after configured duration
//
// Idempotency:
//   - Not idempotent - each registration creates a new user
//   - Duplicate emails return Conflict error
//
// Token Strategy:
//   - Access Token: Short-lived JWT for API authorization
//   - Refresh Token: Long-lived token for obtaining new access tokens
//   - Both tokens returned to client immediately after registration
//   - Refresh token is hashed in database (prevents token theft if DB compromised)
//
// Device Fingerprinting:
//   - User agent: Browser/app identifier
//   - IP address: Client network location
//   - Device ID: Stable client-generated identifier
//   - Enables security features like suspicious login detection
//
// Error Responses:
//   - Validation:      Invalid email format, weak password, missing name fields
//   - Conflict:        Email already exists
//   - Internal:        Database error, token generation failure, hashing failure
//
// Example:
//
//	resp, err := service.Register(ctx, &authv1.RegisterRequest{
//	    Email:     "user@example.com",
//	    Password:  "securePassword123!",
//	    FirstName: "John",
//	    LastName:  "Doe",
//	}, clientInfo)
//	if err != nil {
//	    return nil, err
//	}
//	fmt.Printf("Access token: %s", resp.AccessToken)
//	fmt.Printf("Refresh token: %s", resp.RefreshToken)
func (s *AuthService) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	// Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("register request is required")
	}

	// Extract client information for device fingerprinting
	clientInfo, ok := metadata.ReceiveClientInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "client info not found in context")
	}

	// Check if email already exists (early validation to prevent unnecessary work)
	exists, err := s.DBQ.CheckUserEmailExists(ctx, req.Email)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to check email existence").
			WithDetail("db_error", err.Error())
	}
	if exists {
		return nil, apperror.ErrConflict.WithMessage("email already exists")
	}

	// Hash the password
	hashedPassword, appErr := s.HashService.Hash(req.GetPassword())
	if appErr != nil {
		return nil, appErr
	}

	// Store results for use after transaction
	var (
		user             *db.User
		profile          *db.Profile
		refreshTokenHash string
		accessToken      string
	)

	// Execute registration in transaction
	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Create user record
		u, err := q.CreateUser(ctx, db.CreateUserParams{
			Email:        req.GetEmail(),
			PasswordHash: pgtype.Text{String: hashedPassword, Valid: true},
		})
		if err != nil {
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to create user record").
				WithDetail("db_error", err.Error())
		}
		user = &u

		// Create user profile
		p, appErr := s.createProfile(ctx, q, user.ID, req)
		if appErr != nil {
			return appErr
		}
		profile = p

		// Create session
		session, appErr := s.createSession(ctx, q, user.ID, refreshTokenHash, &clientInfo)
		if appErr != nil {
			return appErr
		}

		// Issue access token and refresh token
		tokenPair, appErr := s.issueTokenPair(ctx, user.ID.String(), session.ID.String())
		if appErr != nil {
			return appErr
		}
		accessToken = tokenPair.AccessToken
		refreshTokenHash = tokenPair.RefreshTokenHash

		// Log successful registration
		s.Logger.Info("new user registered successfully",
			zap.String("user_id", user.ID.String()),
			zap.String("email", user.Email))

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// TODO: Send verification email
	if _, err = s.RequestEmailVerification(ctx, &authv1.RequestEmailVerificationRequest{
		Email: user.Email,
	}); err != nil {
		s.Logger.Error("failed to send verification email",
			zap.Error(err))
	}

	// Build and return authentication response
	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

// Logout terminates the user's current session and revokes authentication tokens.
//
// This method handles user logout by revoking the access token and invalidating
// the session, preventing further use of the same authentication credentials.
//
// Security Constraints:
//   - Access token is immediately revoked (cannot be used for new requests)
//   - Session is marked as revoked in database
//   - Refresh token cannot be used to issue new access tokens
//   - Idempotent operation - calling logout multiple times has no negative effect
//
// Idempotency:
//   - Safe to call multiple times (already revoked sessions return success)
//   - Revoking an already revoked token/session is a no-op
//
// Error Responses:
//   - Internal: Failed to parse session ID, token revocation failed, or database error
//   - PermissionDenied: User not authenticated (no user info in context)
//
// Example:
//
//	success, err := service.Logout(ctx, &corev1.EmptyRequest{})
//	if err != nil {
//	    return nil, err
//	}
//	fmt.Printf("Logged out: %v", success)
func (s *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*corev1.SuccessResponse, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}
	sessionID, err := uuid.Parse(userInfo.SessionID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to parse session ID").WithDetail("error", err.Error())
	}

	// Extract client information for device fingerprinting
	clientInfo, ok := metadata.ReceiveClientInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "client info not found in context")
	}

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		session, err := q.GetSession(ctx, sessionID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrNotFound.WithMessage("session not found")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to get session by id").
				WithDetail("db_error", err.Error())
		}
		if session.DeviceID != clientInfo.DeviceID {
			return apperror.ErrPermissionDenied.WithMessage("device mismatch")
		}

		if err := q.RevokeSession(ctx, session.ID); err != nil {
			return apperror.ErrInternal.WithDetail("internal_message", "failed to revoke session").WithDetail("db_error", err.Error())
		}

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// Log successful logout for audit trail
	s.Logger.Info("user logged out successfully",
		zap.String("user_id", userInfo.UserID),
		zap.String("session_id", userInfo.SessionID))

	return &corev1.SuccessResponse{Success: true}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshAccessTokenRequest) (*authv1.RefreshTokenResponse, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	// Extract client information for device fingerprinting
	clientInfo, ok := metadata.ReceiveClientInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "client info not found in context")
	}

	var (
		accessToken      string
		refreshTokenHash string
	)

	if appErr := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Validate access token with client info (DeviceID)
		session, err := q.GetSessionByRefreshTokenHash(ctx, req.RefreshToken)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrNotFound.WithMessage("session not found")
			}
			return apperror.ErrInternal.
				WithDetail("internal_message", "failed to get session by refresh token hash").
				WithDetail("db_error", err.Error())
		}

		if session.DeviceID != clientInfo.DeviceID {
			return apperror.ErrPermissionDenied.WithMessage("device mismatch")
		}

		// Issue new access token and refresh token
		tokenPair, appErr := s.issueTokenPair(ctx, userInfo.UserID, userInfo.SessionID)
		if appErr != nil {
			return appErr
		}
		accessToken = tokenPair.AccessToken
		refreshTokenHash = tokenPair.RefreshTokenHash

		// Persist tokens in db
		q.UpdateSessionRefreshTokenHash(ctx, db.UpdateSessionRefreshTokenHashParams{
			ID:               session.ID,
			RefreshTokenHash: refreshTokenHash,
		})

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	return &authv1.RefreshTokenResponse{
		AccessToken: &modelsv1.Token{
			Value:     accessToken,
			ExpiresAt: coreutils.ParseToProtoTimestamp(s.Config.SessionTTL),
		},
		RefreshToken: refreshTokenHash,
	}, nil
}

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

		p, appErr := s.getUserProfile(ctx, q, user.ID)
		if appErr != nil {
			return appErr
		}
		profile = p

		scopedToken, appErr := s.TokenManager.IssueScopedToken(ctx, user.ID.String(), constants.TokenScopeVerifyEmail)
		if appErr != nil {
			return appErr
		}

		url, appErr := utils.NewTokenURL(scopedToken.TokenString, s.Config.FrontendURL, s.Config.EmailVerificationPath)
		if appErr != nil {
			return appErr
		}
		verificationURL = url

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	if appErr := s.OrgOpenFGAPublisher.Publish(jobs.JobUserRequestedEmailVerification, dto.EmailVerificationDTO{
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

		p, appErr := s.getUserProfile(ctx, q, user.ID)
		if appErr != nil {
			return appErr
		}
		profile = p

		scopedToken, appErr := s.TokenManager.IssueScopedToken(ctx, user.ID.String(), constants.TokenScopeVerifyEmail)
		if appErr != nil {
			return appErr
		}

		url, appErr := utils.NewTokenURL(scopedToken.TokenString, s.Config.FrontendURL, s.Config.EmailVerificationPath)
		if appErr != nil {
			return appErr
		}
		resetURL = url

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	if appErr := s.OrgOpenFGAPublisher.Publish(jobs.JobUserRequestedPasswordReset, dto.PasswordResetDTO{
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
