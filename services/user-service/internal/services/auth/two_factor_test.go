package auth_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/testutils"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: add edge cases tests

func Test_InitTwoFactorTOTP(t *testing.T) {
	// Core Utils
	ctx := context.Background()

	// User credentials
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)

	// Create a test user
	user, _, err := createUser(ctx, email, password)
	require.NoError(t, err)

	// Enable two-factor authentication
	// Set user info in the context
	ctx = setUserInfoInCtx(ctx, user)
	res, err := authService.InitTwoFactorTOTP(ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, res)
	// Check that the response contains the expected fields
	assert.NotEmpty(t, res.TwoFactorSecret)
	assert.NotEmpty(t, res.QrCodeUri)
}

func Test_EnableTwoFactorTOTP(t *testing.T) {
	// Core Utils
	ctx := context.Background()

	// User credentials
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)

	// Create a test user
	user, _, err := createUser(ctx, email, password)
	require.NoError(t, err)
	ctx = setUserInfoInCtx(ctx, user) // Set user info in the context

	// Initialize two-factor authentication
	res, err := authService.InitTwoFactorTOTP(ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Generate a valid TOTP
	totp, appErr := generate2FATokenCode(res.GetTwoFactorSecret())
	require.Nil(t, appErr)

	// Test Enable method
	successRes, err := authService.EnableTwoFactorTOTP(ctx, &authv1.TwoFactorTOTPRequest{
		Totp:            totp,
		TwoFactorSecret: res.TwoFactorSecret,
	})
	require.NoError(t, err)
	require.NotNil(t, successRes)

	assert.True(t, successRes.Success)

}

func Test_DisableTwoFactor(t *testing.T) {
	ctx := context.Background()

	// User Credentials
	email := testutils.GenerateRandomEmail()
	password := testutils.GenerateRandomString(10)

	// Create a user for testing
	user, _, err := createUser(ctx, email, password)
	t.Cleanup(func() {
		authService.DBQ.DeleteUserHard(context.Background(), user.ID)
	})
	require.NoError(t, err)
	ctx = setUserInfoInCtx(ctx, user) // Set user info in the context

	// Init two factor authentication
	res, err := authService.InitTwoFactorTOTP(ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Generate TOTP code
	totp, appErr := generate2FATokenCode(res.GetTwoFactorSecret())
	require.Nil(t, appErr)

	// Enable two-factor authentication
	_, err = authService.EnableTwoFactorTOTP(ctx, &authv1.TwoFactorTOTPRequest{
		Totp:            totp,
		TwoFactorSecret: res.TwoFactorSecret,
	})
	require.NoError(t, err)

	// Disable two-factor authentication
	successRes, err := authService.DisableTwoFactor(ctx, &corev1.EmptyRequest{})
	require.NoError(t, err)
	require.NotNil(t, successRes)
	require.True(t, successRes.Success)

	// Try to login
	// Should directly pass without two-factor authentication
	_, err = authService.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
}

// =======================================================
// Helper functions
// =======================================================

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
