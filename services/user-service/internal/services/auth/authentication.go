package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/metadata"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/utils"
	"go.uber.org/zap"
)

// Login authenticates a user with email and password, creating a new session.
//
// Business Rile:
//   - If user has enabled 2FA then after authenticating the user, issue a new pre auth token, which will act as a authentication token (only allowed to LoginWithTwoFactorCode method)
//
// Idempotent:
//   - Not idempotent. Each call creates a new session.
//
// Constraints:
//   - Request must not be nil (validation of the req parameters is done by gateway)
//   - Verify the password with the hashed password in the database
//
// Race condition:
//   - If the issueAuthToken fails, there will be a session created but not used by the user. These need to be cleaned up by daily cron jobs.
//
// TODO: implement user with 2FA
func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	// Validate request
	if req == nil {
		return nil, apperror.ErrInternal.WithMessage("login request is required")
	}

	// Extract client information for device fingerprinting
	clientInfo, ok := metadata.GetClientInfoFromContext(ctx)
	if !ok {
		return nil, constants.ErrClientNotFoundInCtx
	}

	// Retrieve user by email
	user, err := s.DBQ.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Generic error prevents user enumeration
			return nil, apperror.ErrPermissionDenied.WithMessage("invalid credentials")
		}
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to get user by email").
			WithDetail("db_error", err.Error())
	}

	// Verify password
	if !s.HashService.Verify(user.PasswordHash.String, req.GetPassword()) {
		return nil, apperror.ErrPermissionDenied.WithMessage("invalid credentials")
	}

	exists, err := s.DBQ.CheckTwoFactorAuthEnabledByUserID(ctx, user.ID)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to get primary two-factor auth by user id").
			WithDetail("db_error", err.Error())
	}
	if exists {
		// If 2FA is enabled, issue a pre-auth token and return 2FA required status
		preAuthToken, appErr := s.issuePreAuthToken(ctx, user.ID.String())
		if appErr != nil {
			return nil, appErr
		}
		return &authv1.AuthResponse{
			Status:       authv1.AuthStatus_AUTH_STATUS_2FA_REQUIRED,
			PreAuthToken: preAuthToken,
		}, nil
	}

	// Retrieve user profile
	profile, err := s.DBQ.GetProfileByUserID(ctx, user.ID)
	if appErr := utils.AssertRowExists(err, "profile", user.ID.String()); appErr != nil {
		return nil, appErr
	}

	_, authTokens, appErr := s.crateSessionAndIssueTokens(ctx, user.ID, &clientInfo)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapAuthResponse(&user, &profile, authTokens.AccessToken, authTokens.RefreshToken), nil
}

// TODO: create a function for oauth user login and registration

// Register creates a new user account and initiates the authentication flow.
//
// Idempotency:
//   - Not idempotent - each registration creates a new user
//   - Duplicate emails return Conflict error
//
// Race condition:
//   - If the issueAuthToken fails, there will be a session created but not used by the user. These need to be cleaned up by daily cron jobs.
//
// Error Responses:
//   - Validation:      Invalid email format, weak password, missing name fields
//   - Conflict:        Email already exists
//   - Internal:        Database error, token generation failure, hashing failure
func (s *AuthService) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	// Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("register request is required")
	}

	// Extract client information for device fingerprinting
	clientInfo, ok := metadata.GetClientInfoFromContext(ctx)
	if !ok {
		return nil, constants.ErrClientNotFoundInCtx
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
		session          *db.Session
		profile          *db.Profile
		refreshTokenHash string
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

		// Create session along with a refresh token hash
		sess, appErr := s.createSession(ctx, q, user.ID, &clientInfo)
		if appErr != nil {
			return appErr
		}
		session = sess
		refreshTokenHash = session.RefreshTokenHash

		return nil
	}); appErr != nil {
		return nil, appErr
	}

	// Issue access token
	accessToken, appErr := s.issueAccessToken(ctx, user.ID.String(), session.ID.String())
	if appErr != nil {
		return nil, appErr
	}

	// Log successful registration
	s.Logger.Info("new user registered successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email))

	// Send verification email
	if _, err = s.RequestEmailVerification(ctx, &authv1.RequestEmailVerificationRequest{
		Email: user.Email,
	}); err != nil {
		s.Logger.Error("failed to send verification email",
			zap.Error(err))
	}

	// Map and return authentication response
	return utils.MapAuthResponse(user, profile, accessToken, refreshTokenHash), nil
}

func (s *AuthService) LoginWithTwoFactorCode(ctx context.Context, req *authv1.TwoFactorCodeRequest) (*authv1.AuthResponse, error) {
	// Validate request parameters
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("token request is required")
	}

	// Extract client information for device fingerprinting
	clientInfo, ok := metadata.GetClientInfoFromContext(ctx)
	if !ok {
		return nil, constants.ErrClientNotFoundInCtx
	}

	// Extract tokens information from authenticated context
	scopedTokenInfo, ok := metadata.GetScopedTokenInfoFromContext(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve scoped token info from context")
	}
	// Validate pre auth token
	if appErr := validate2FAToken(&scopedTokenInfo); appErr != nil {
		return nil, appErr
	}
	// Validate two factor code
	if appErr := validate2FACode(req.GetTwoFactorCode()); appErr != nil {
		return nil, appErr
	}

	// Get User from database
	userID, err := uuid.Parse(scopedTokenInfo.Subject)
	if err != nil {
		return nil, constants.ErrInvalidUserIDInUserInfo
	}
	user2FA, err := s.DBQ.GetTwoFactorAuthByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user 2FA from database")
	}
	// Generate and Match two factor code against generated code
	generatedTwoFactorCode, appErr := generate2FATokenCode(user2FA.Secret)
	if appErr != nil {
		return nil, appErr
	}
	if req.GetTwoFactorCode() != generatedTwoFactorCode {
		return nil, apperror.ErrValidation.WithMessage("invalid two factor code")
	}

	// Fetch user from database
	user, err := s.DBQ.GetUser(ctx, userID)
	if appErr := utils.AssertRowExists(err, "user", userID.String()); appErr != nil {
		return nil, appErr
	}

	// Fetch profile from database
	profile, err := s.DBQ.GetProfileByUserID(ctx, userID)
	if appErr := utils.AssertRowExists(err, "profile", scopedTokenInfo.Subject); appErr != nil {
		return nil, appErr
	}

	// Create session and issue access and refresh tokens
	_, authTokens, appErr := s.crateSessionAndIssueTokens(ctx, userID, &clientInfo)
	if appErr != nil {
		return nil, appErr
	}

	// Return authentication response
	return utils.MapAuthResponse(&user, &profile, authTokens.AccessToken, authTokens.RefreshToken), nil
}

// ================================================================
// CORE HELPER FUNCTIONS
// ===============================================================

func (s *AuthService) crateSessionAndIssueTokens(ctx context.Context, userID uuid.UUID, clientInfo *dto.ClientInfo) (*db.Session, *dto.AuthTokens, *apperror.AppError) {
	var session *db.Session
	var refreshTokenHash string

	// Execute authentication in transaction
	if err := s.Helper.RunInTx(ctx, func(q *db.Queries) *apperror.AppError {
		// Create session along with refresh token
		s, appErr := s.createSession(ctx, q, userID, clientInfo)
		if appErr != nil {
			return appErr
		}
		session = s
		refreshTokenHash = session.RefreshTokenHash

		return nil
	}); err != nil {
		return nil, nil, err
	}

	// Issue access token
	token, appErr := s.issueAccessToken(ctx, userID.String(), session.ID.String())
	if appErr != nil {
		return nil, nil, appErr
	}

	return session, &dto.AuthTokens{
		AccessToken:  token,
		RefreshToken: refreshTokenHash,
	}, nil
}

func (s *AuthService) issuePreAuthToken(ctx context.Context, userID string) (string, *apperror.AppError) {
	tokenRes, appErr := s.TokenManager.IssueScopedToken(ctx, userID, constants.TokenScope2FA)
	if appErr != nil {
		return "", appErr
	}

	return tokenRes.TokenString, nil
}

func (s *AuthService) createSession(
	ctx context.Context, q *db.Queries,
	userID uuid.UUID, clientInfo *dto.ClientInfo,
) (*db.Session, *apperror.AppError) {
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(s.Config.SessionTTL),
		Valid: true,
	}

	hash, appErr := s.HashService.Generate(32)
	if appErr != nil {
		return nil, appErr
	}

	session, err := q.CreateSession(ctx, db.CreateSessionParams{
		UserID:           userID,
		UserAgent:        clientInfo.UserAgent,
		IpAddr:           clientInfo.IPAddress,
		DeviceID:         clientInfo.DeviceID,
		ExpiresAt:        expiresAt,
		RefreshTokenHash: hash,
	})
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to create session").
			WithDetail("db_error", err.Error())
	}

	return &session, nil
}

// createProfile creates a user profile record.
func (s *AuthService) createProfile(ctx context.Context, q *db.Queries, userID uuid.UUID, req *authv1.RegisterRequest) (*db.Profile, *apperror.AppError) {
	profile, err := q.CreateProfile(ctx, db.CreateProfileParams{
		UserID:    userID,
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	})
	if err != nil {
		return nil, apperror.ErrInternal.
			WithDetail("internal_message", "failed to create user profile").
			WithDetail("db_error", err.Error())
	}

	return &profile, nil
}

func (s *AuthService) issueAccessToken(ctx context.Context, userID, sessionID string) (string, *apperror.AppError) {
	accessTokenRes, appErr := s.TokenManager.IssueAuthToken(ctx, userID, sessionID)
	if appErr != nil {
		return "", appErr
	}

	return accessTokenRes.TokenString, nil
}

// ================================================================
// CORE VALIDATION HELPER FUNCTIONS
// ================================================================

func validate2FACode(code string) *apperror.AppError {
	if code == "" {
		return apperror.ErrValidation.WithMessage("code is required")
	}
	if len(code) != 6 {
		return apperror.ErrValidation.WithMessage("code must be 6 digits")
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return apperror.ErrValidation.WithMessage("code must be numeric")
		}
	}
	return nil
}

func generate2FATokenCode(secret string) (string, *apperror.AppError) {
	// 1. Validate input
	if secret == "" {
		return "", apperror.ErrValidation.WithMessage("secret is required")
	}

	// 2. Decode base32 secret (authenticator apps use base32)
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", apperror.ErrInternal.WithMessage("invalid secret format")
	}

	// 3. Calculate current time step (30-second intervals)
	timeStep := uint64(math.Floor(float64(time.Now().Unix()) / 30))

	// 4. Convert timeStep to 8-byte big-endian bytes
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, timeStep)

	// 5. Generate HMAC-SHA1
	h := hmac.New(sha1.New, key)
	h.Write(timeBytes)
	hash := h.Sum(nil) // 20 bytes

	// 6. Dynamic truncation (RFC 4226)
	offset := hash[len(hash)-1] & 0x0f
	truncatedHash := hash[offset : offset+4]

	// 7. Convert to 31-bit integer and get 6-digit code
	code := (binary.BigEndian.Uint32(truncatedHash) & 0x7fffffff) % 1000000

	// 8. Return zero-padded 6-digit string
	return fmt.Sprintf("%06d", code), nil
}

// ================================================================
// SIMPLE VALIDATION HELPER FUNCTIONS
// ================================================================

func validate2FAToken(tokensInfo *dto.ScopedToken) *apperror.AppError {
	if tokensInfo.Scope != constants.TokenScope2FA {
		return apperror.ErrValidation.WithMessage("invalid preauth token scope")
	}
	return nil
}
