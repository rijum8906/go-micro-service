package auth_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
	mock_token "github.com/rijum8906/relay/packages/core/token/mocks"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	authService      *auth.AuthService
	logger           *zap.Logger
	tokenManager     token.TokenManager
	mockTokenManager *mock_token.MockTokenManager
)

func TestMain(m *testing.M) {
	apperror.SetConfig(&apperror.Config{
		Logger: zap.NewNop(),
		AppEnv: "test",
		Debug:  true,
	})

	authService = auth.NewForTest()

	logger = authService.Logger
	tokenManager = authService.TokenManager

	mockTokenManager = &mock_token.MockTokenManager{}

	exitCode := m.Run()
	os.Exit(exitCode)
}

// =====================================================
// Test Suite
// =====================================================

type AuthTestSuite struct {
	ctx context.Context
}

type TestCase[T any] struct {
	name     string
	setupCtx func(ctx context.Context) context.Context
	req      *T
	wantErr  bool
	errCode  apperror.ErrorCode
	contains string
}

// =====================================================
// LOGIN TESTS
// =====================================================

func Test_Login_Validation(t *testing.T) {
	baseCtx := context.Background()

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	user, _, err := createUser(baseCtx, email, password)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := authService.DBQ.DeleteUserHard(baseCtx, user.ID); err != nil {
			logger.Error("failed to cleanup user", zap.Error(err))
		}
	})

	testCases := []TestCase[authv1.LoginRequest]{
		{
			name: "nil req",
			setupCtx: func(ctx context.Context) context.Context {
				return ctx
			},
			req:      nil,
			wantErr:  true,
			errCode:  apperror.CodeInternal,
			contains: "request is required",
		},
		{
			name: "invalid context without client info",
			setupCtx: func(ctx context.Context) context.Context {
				return ctx
			},
			req:      &authv1.LoginRequest{},
			wantErr:  true,
			errCode:  apperror.CodeInternal,
			contains: constants.ErrClientNotFoundInCtx.Message,
		},
		{
			name: "email does not exists",
			setupCtx: func(ctx context.Context) context.Context {
				return getClientInfoBasedCtx(ctx)
			},
			req: &authv1.LoginRequest{
				Email:    testutils.GenerateRandomEmail(),
				Password: testutils.GenerateRandomString(32),
			},
			wantErr:  true,
			errCode:  apperror.CodePermissionDenied,
			contains: "invalid credentials",
		},
		{
			name: "Invalida password",
			setupCtx: func(ctx context.Context) context.Context {
				return getClientInfoBasedCtx(ctx)
			},
			req: &authv1.LoginRequest{
				Email:    user.Email,
				Password: testutils.GenerateRandomString(32),
			},
			wantErr:  true,
			errCode:  apperror.CodePermissionDenied,
			contains: "invalid credentials",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx(baseCtx)
			_, err := authService.Login(ctx, tc.req)

			if tc.wantErr {
				appErr, ok := err.(*apperror.AppError)
				require.True(t, ok)

				assert.NotNil(t, appErr)
				assert.Equal(t, string(tc.errCode), string(appErr.Code))

				assert.Contains(t, appErr.Message, tc.contains)

				// Log to test output instead of global logger to avoid duplicate internal logs
				logger.Error("test failed", apperror.ParseAppErrorIntoZapFields(appErr)...)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_Login_Edge_Cases(t *testing.T) {
	ctx := getClientInfoBasedCtx(context.Background())

	email1 := testutils.GenerateRandomEmail()
	password1 := testutils.GenerateRandomString(32)
	email2 := testutils.GenerateRandomEmail()
	password2 := testutils.GenerateRandomString(32)

	// User 1
	user1, _, err := createUser(ctx, email1, password1)
	if err != nil {
		logger.Error("failed to create user", zap.Error(err))
		t.FailNow()
	}
	t.Cleanup(func() {
		if err := authService.DBQ.DeleteUserHard(ctx, user1.ID); err != nil {
			logger.Error("failed to cleanup user", zap.Error(err))
		}
	})

	// User 2
	user2, profile2, err := createUser(ctx, email2, password2)
	if err != nil {
		logger.Error("failed to create user", zap.Error(err))
		t.FailNow()
	}
	t.Cleanup(func() {
		if err := authService.DBQ.DeleteUserHard(ctx, user2.ID); err != nil {
			logger.Error("failed to cleanup user", zap.Error(err))
		}
	})

	// User without a profile
	t.Run("user without a profile", func(t *testing.T) {
		if err := authService.DBQ.DeleteProfileHard(ctx, profile2.ID); err != nil {
			logger.Error("failed to delete profile", zap.Error(err))
			t.FailNow()
		}

		_, err := authService.Login(ctx, &authv1.LoginRequest{
			Email:    email2,
			Password: password2,
		})

		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)

		require.Equal(t, appErr.Code, apperror.CodeInternal)
		require.Contains(t, appErr.Message, "profile not found")
	})

	// Redis server error
	t.Run("redis server closed", func(t *testing.T) {
		mockTokenManager.On("IssueAuthToken", ctx, user1.ID.String(), mock.Anything).
			Return(nil, apperror.ErrInternal)

		authService.TokenManager = mockTokenManager
		t.Cleanup(func() {
			authService.TokenManager = tokenManager
		})

		_, err = authService.Login(ctx, &authv1.LoginRequest{
			Email:    email1,
			Password: password1,
		})

		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)

		require.Equal(t, appErr.Code, apperror.CodeInternal)
	})
}

func Test_Login_Success(t *testing.T) {
	ctx := getClientInfoBasedCtx(context.Background())

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	user, profile, err := createUser(ctx, email, password)
	if err != nil {
		logger.Error("failed to create user", zap.Error(err))
		t.FailNow()
	}

	res, err := authService.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})

	require.NoError(t, err)

	assert.Equal(t, user.ID.String(), res.GetUser().GetId())
	assert.Equal(t, profile.ID.String(), res.GetProfile().GetId())

	// cleanup
	t.Cleanup(func() {
		if err := authService.DBQ.DeleteUserHard(context.Background(), user.ID); err != nil {
			logger.Error("failed to cleanup user", zap.Error(err))
		}
	})
}

// ======================================================
// REGISTER TESTS
// ======================================================

func Test_Register_Validation(t *testing.T) {
	baseCtx := context.Background()

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	user, _, err := createUser(baseCtx, email, password)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := authService.DBQ.DeleteUserHard(baseCtx, user.ID); err != nil {
			logger.Error("failed to cleanup user", zap.Error(err))
		}
	})

	testCases := []TestCase[authv1.RegisterRequest]{
		{
			name: "nil req",
			setupCtx: func(ctx context.Context) context.Context {
				return ctx
			},
			req:      nil,
			wantErr:  true,
			errCode:  apperror.CodeValidation,
			contains: "register request is required",
		},
		{
			name: "invalid context without client info",
			setupCtx: func(ctx context.Context) context.Context {
				return ctx
			},
			req:      &authv1.RegisterRequest{},
			wantErr:  true,
			errCode:  apperror.CodeInternal,
			contains: constants.ErrClientNotFoundInCtx.Message,
		},
		{
			name: "email already exists",
			setupCtx: func(ctx context.Context) context.Context {
				return getClientInfoBasedCtx(ctx)
			},
			req: &authv1.RegisterRequest{
				Email:     user.Email,
				Password:  testutils.GenerateRandomString(32),
				FirstName: testutils.GenerateRandomString(6),
				LastName:  testutils.GenerateRandomString(6),
			},
			wantErr:  true,
			errCode:  apperror.CodeConflict,
			contains: "email already exists",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx(baseCtx)
			_, err := authService.Register(ctx, tc.req)

			if tc.wantErr {
				appErr, ok := err.(*apperror.AppError)
				require.True(t, ok)

				assert.NotNil(t, appErr)
				assert.Equal(t, string(tc.errCode), string(appErr.Code))

				assert.Contains(t, appErr.Message, tc.contains)

				// Log to test output instead of global logger to avoid duplicate internal logs
				logger.Error("test failed", apperror.ParseAppErrorIntoZapFields(appErr)...)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_Register_Edge_Cases(t *testing.T) {
	ctx := getClientInfoBasedCtx(context.Background())

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	// Redis server error when issuing token
	t.Run("redis server closed", func(t *testing.T) {
		mockTokenManager.On("IssueAuthToken", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, apperror.ErrInternal)

		authService.TokenManager = mockTokenManager
		t.Cleanup(func() {
			authService.TokenManager = tokenManager
		})

		_, err := authService.Register(ctx, &authv1.RegisterRequest{
			Email:     email,
			Password:  password,
			FirstName: testutils.GenerateRandomString(6),
			LastName:  testutils.GenerateRandomString(6),
		})

		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)

		require.Equal(t, appErr.Code, apperror.CodeInternal)
	})
}

func Test_Register_Success(t *testing.T) {
	ctx := getClientInfoBasedCtx(context.Background())

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	resp, err := authService.Register(ctx, &authv1.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: testutils.GenerateRandomString(6),
		LastName:  testutils.GenerateRandomString(6),
	})

	require.NoError(t, err)

	userID, parseErr := uuid.Parse(resp.GetUser().GetId())
	require.NoError(t, parseErr)

	// cleanup
	t.Cleanup(func() {
		if err := authService.DBQ.DeleteUserHard(context.Background(), userID); err != nil {
			logger.Error("failed to cleanup user", zap.Error(err))
		}
	})
}

// ======================================================
// Helper functions
// ======================================================

func createUser(ctx context.Context, email, password string) (*db.User, *db.Profile, error) {
	passwordHash, appErr := authService.HashService.Hash(password)
	if appErr != nil {
		return nil, nil, appErr
	}

	// create user
	user, err := authService.DBQ.CreateUser(ctx, db.CreateUserParams{
		Email: email,
		PasswordHash: pgtype.Text{
			String: passwordHash,
			Valid:  true,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	// Create Profile
	profile, err := authService.DBQ.CreateProfile(ctx, db.CreateProfileParams{
		UserID:    user.ID,
		FirstName: testutils.GenerateRandomString(10),
		LastName:  testutils.GenerateRandomString(10),
		AvatarUrl: testutils.GenerateRandomString(40),
	})
	if err != nil {
		return nil, nil, err
	}

	return &user, &profile, nil
}

func getClientInfoBasedCtx(ctx context.Context) context.Context {
	clientInfo := dto.ClientInfo{
		DeviceID:   testutils.GenerateRandomString(32),
		IPAddress:  "127.0.0.1",
		UserAgent:  testutils.GenerateRandomString(32),
		ClientType: "web",
		APIVersion: "0.0.1",
		Locale:     "en-US",
		SDKVersion: "1.0.0",

		RequestID: uuid.NewString(),
		TraceID:   uuid.NewString(),
	}
	return metadata.NewIncomingContext(ctx, metadata.Pairs(
		dto.MetaDeviceIDKey, clientInfo.DeviceID,
		dto.MetaClientIPKey, clientInfo.IPAddress,
		dto.MetaUserAgentKey, clientInfo.UserAgent,
		dto.MetaClientTypeKey, clientInfo.ClientType,
		dto.MetaAPIVersionKey, clientInfo.APIVersion,
		dto.MetaLocaleKey, clientInfo.Locale,
		dto.MetaSDKVersionKey, clientInfo.SDKVersion,
		dto.MetaRequestIDKey, clientInfo.RequestID,
		dto.MetaTraceIDKey, clientInfo.TraceID,
	))
}
