package auth_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	mock_broker "github.com/rijum8906/relay/packages/core/broker/mocks"
	"github.com/rijum8906/relay/packages/core/corelogger"
	"github.com/rijum8906/relay/packages/core/testutils"
	mock_token "github.com/rijum8906/relay/packages/core/token/mocks"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/app/constants"
	"github.com/rijum8906/relay/services/user/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	apperror.SetConfig(&apperror.Config{
		Logger: corelogger.NewDevLogger(),
		AppEnv: "test",
		Debug:  true,
	})

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
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := suite.AuthService.DBQ.DeleteUserHard(suite.Ctx, user.ID); err != nil {
			suite.AuthService.Logger.Error("failed to cleanup user", zap.Error(err))
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
				suite.SetClientInfoInContext()
				return suite.Ctx
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
				suite.SetClientInfoInContext()
				return suite.Ctx
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
			ctx := tc.setupCtx(suite.Ctx)
			_, err := suite.AuthService.Login(ctx, tc.req)

			if tc.wantErr {
				appErr, ok := err.(*apperror.AppError)
				require.True(t, ok)

				assert.NotNil(t, appErr)
				assert.Equal(t, string(tc.errCode), string(appErr.Code))

				assert.Contains(t, appErr.Message, tc.contains)

				// Log to test output instead of global logger to avoid duplicate internal logs
				suite.AuthService.Logger.Error("test failed", apperror.ParseAppErrorIntoZapFields(appErr)...)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_Login_Edge_Cases(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	suite.SetClientInfoInContext()
	ctx := suite.Ctx

	email1 := testutils.GenerateRandomEmail()
	password1 := testutils.GenerateRandomString(32)
	email2 := testutils.GenerateRandomEmail()
	password2 := testutils.GenerateRandomString(32)

	// User 1
	user1, _, err := suite.CreateUser(t, email1, password1)
	if err != nil {
		suite.AuthService.Logger.Error("failed to create user", zap.Error(err))
		t.FailNow()
	}
	t.Cleanup(func() {
		if err := suite.AuthService.DBQ.DeleteUserHard(ctx, user1.ID); err != nil {
			suite.AuthService.Logger.Error("failed to cleanup user", zap.Error(err))
		}
	})

	// User 2
	user2, profile2, err := suite.CreateUser(t, email2, password2)
	if err != nil {
		suite.AuthService.Logger.Error("failed to create user", zap.Error(err))
		t.FailNow()
	}
	t.Cleanup(func() {
		if err := suite.AuthService.DBQ.DeleteUserHard(ctx, user2.ID); err != nil {
			suite.AuthService.Logger.Error("failed to cleanup user", zap.Error(err))
		}
	})

	// User without a profile
	t.Run("user without a profile", func(t *testing.T) {
		if err := suite.AuthService.DBQ.DeleteProfileHard(ctx, profile2.ID); err != nil {
			suite.AuthService.Logger.Error("failed to delete profile", zap.Error(err))
			t.FailNow()
		}

		_, err := suite.AuthService.Login(ctx, &authv1.LoginRequest{
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
		mockTokenManager := mock_token.NewMockTokenManager(t)
		mockTokenManager.On("IssueAuthToken", ctx, user1.ID.String(), mock.Anything).
			Return(nil, apperror.ErrInternal)

		suite.AuthService.TokenManager = mockTokenManager

		_, err = suite.AuthService.Login(ctx, &authv1.LoginRequest{
			Email:    email1,
			Password: password1,
		})

		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)

		require.Equal(t, appErr.Code, apperror.CodeInternal)
	})
}

func Test_Login_Success_Without_2FA(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	suite.SetClientInfoInContext()
	ctx := suite.Ctx

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	user, profile, err := suite.CreateUser(t, email, password)
	if err != nil {
		suite.AuthService.Logger.Error("failed to create user", zap.Error(err))
		t.FailNow()
	}

	res, err := suite.AuthService.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})

	require.NoError(t, err)

	assert.Equal(t, user.ID.String(), res.GetUser().GetId())
	assert.Equal(t, profile.ID.String(), res.GetProfile().GetId())

	// cleanup
	t.Cleanup(func() {
		if err := suite.AuthService.DBQ.DeleteUserHard(context.Background(), user.ID); err != nil {
			suite.AuthService.Logger.Error("failed to cleanup user", zap.Error(err))
		}
	})
}

// Test_Login_Success_With_2FA tests a successful login with 2FA enabled
//
// Preconditions:
// - Must tests all the functions involved in enabling 2FA
func Test_Login_Success_With_2FA(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	suite.SetClientInfoInContext()

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	// 1. Create a user
	user, _, err := suite.CreateUser(t, email, password)
	if err != nil {
		suite.AuthService.Logger.Error("failed to create user", zap.Error(err))
		t.FailNow()
	}
	suite.SetUserInfoInContext(user) // InitTwoFactorTOTP and EnableTwoFactorTOTP require authenticated context

	// 2. Enable 2FA for the user
	res, err := suite.AuthService.InitTwoFactorTOTP(suite.Ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, res.TwoFactorSecret)
	require.NotEmpty(t, res.QrCodeUri)

	// Generate a valid TOTP code
	totpCode, appErr := generate2FATokenCode(res.TwoFactorSecret)
	require.Nil(t, appErr)

	// Enable 2FA for the user
	successRes, err := suite.AuthService.EnableTwoFactorTOTP(suite.Ctx, &authv1.TwoFactorTOTPRequest{
		Totp:            totpCode,
		TwoFactorSecret: res.TwoFactorSecret,
	})
	require.NoError(t, err)
	require.True(t, successRes.Success)

	// Verify that 2FA is enabled for the user
	r, err := suite.AuthService.DBQ.GetPrimaryTwoFactorAuthByUserID(suite.Ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, r)

	// Try to login with the user's credentials
	loginRes, err := suite.AuthService.Login(suite.Ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	require.Equal(t, authv1.AuthStatus_AUTH_STATUS_2FA_REQUIRED.String(), loginRes.Status.String())

	// Attach a scoped token to the context
	suite.SetScopedTokenInfoInContext(user.ID.String())

	// Now complete the 2FA login
	authRes, err := suite.AuthService.LoginWithTwoFactorCode(suite.Ctx, &authv1.TwoFactorCodeRequest{
		TwoFactorCode: totpCode,
	})
	require.NoError(t, err)
	require.NotNil(t, authRes)

	t.Cleanup(func() {
		if err := suite.AuthService.DBQ.DeleteUserHard(context.Background(), user.ID); err != nil {
			suite.AuthService.Logger.Error("failed to cleanup user", zap.Error(err))
		}
	})
}

// ======================================================
// REGISTER TESTS
// ======================================================

func Test_Register_Validation(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	user, _, err := suite.CreateUser(t, email, password)
	require.NoError(t, err)

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
				suite.SetClientInfoInContext()
				return suite.Ctx
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
			ctx := tc.setupCtx(suite.Ctx)
			_, err := suite.AuthService.Register(ctx, tc.req)

			if tc.wantErr {
				appErr, ok := err.(*apperror.AppError)
				require.True(t, ok)

				assert.NotNil(t, appErr)
				assert.Equal(t, string(tc.errCode), string(appErr.Code))

				assert.Contains(t, appErr.Message, tc.contains)

				// Log to test output instead of global logger to avoid duplicate internal logs
				suite.AuthService.Logger.Error("test failed", apperror.ParseAppErrorIntoZapFields(appErr)...)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_Register_Edge_Cases(t *testing.T) {
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	// Redis server error when issuing token
	t.Run("redis server closed", func(t *testing.T) {
		mockTokenManager := mock_token.NewMockTokenManager(t)
		mockTokenManager.On("IssueAuthToken", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, apperror.ErrInternal)

		suite.AuthService.TokenManager = mockTokenManager

		_, err := suite.AuthService.Register(suite.Ctx, &authv1.RegisterRequest{
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
	suite := auth.NewTestSuite()
	defer suite.TearDownSuite()

	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(32)

	// Attach client info to context
	suite.SetClientInfoInContext()

	// Mock the publisher
	mockPublisher := mock_broker.NewMockPublisher(t)
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	suite.AuthService.OrgOpenFGAPublisher = mockPublisher

	resp, err := suite.AuthService.Register(suite.Ctx, &authv1.RegisterRequest{
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
		if err := suite.AuthService.DBQ.DeleteUserHard(context.Background(), userID); err != nil {
			suite.AuthService.Logger.Error("failed to cleanup user", zap.Error(err))
		}
	})
}
