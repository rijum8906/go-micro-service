package token_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/packages/core/token"
)

type TestCase struct {
	name                string
	tm                  *token.TokenManager // Token Manager to issue with
	tm2                 *token.TokenManager // Token Manager to validate with
	subject             string
	sessionID           string
	scope               token.TokenScope
	wantErrAtIssuing    bool
	errCode1            apperror.ErrorCode
	wantErrAtValidating bool
	errCode2            apperror.ErrorCode
}

func TestTokenManager_IssueValidateAuthToken(t *testing.T) {
	testCases := []TestCase{
		{
			name:             "issue and validate token everything valid",
			tm:               createTokenManager(),
			tm2:              createTokenManager(),
			subject:          testutils.GenerateRandomString(12),
			sessionID:        testutils.GenerateRandomString(12),
			wantErrAtIssuing: false,
		},
		{
			name:                "issue and validate token with invalid jwt secret",
			tm:                  createTokenManager(),
			tm2:                 createTokenManagerWithCustom(&token.Config{JwtSecret: []byte("test_secret_invalid")}),
			subject:             testutils.GenerateRandomString(12),
			sessionID:           testutils.GenerateRandomString(12),
			wantErrAtIssuing:    false,
			wantErrAtValidating: true,
			errCode2:            apperror.CodeTokenInvalidSignature,
		},
		{
			name: "issue and validate with very short expiry time",
			tm: createTokenManagerWithCustom(&token.Config{
				SessionTTL: time.Microsecond,
			}),
			tm2:                 createTokenManager(), // NOTE: using another token manager won't lead to any error
			subject:             testutils.GenerateRandomString(12),
			sessionID:           testutils.GenerateRandomString(12),
			wantErrAtIssuing:    false,
			wantErrAtValidating: true,
			errCode2:            apperror.CodeTokenExpired,
		},
	}

	for _, testCase := range testCases {
		tokenStr, appErr := testCase.tm.IssueAuthToken(context.Background(), testCase.subject, testCase.sessionID, token.TokenScopeAuth)
		if testCase.wantErrAtIssuing {
			if appErr == nil {
				t.Errorf("token validation failed: expected error got nil")
				os.Exit(1)
			}

			if appErr.Code != testCase.errCode1 {
				t.Errorf("token validation failed: expected %v got %v", testCase.errCode1, appErr.Code)
			}
		}

		_, appErr = testCase.tm2.ValidateAuthToken(context.Background(), tokenStr)
		if testCase.wantErrAtValidating {
			if appErr == nil {
				t.Errorf("token validation failed: expected error got nil")
				os.Exit(1)
			}

			if appErr.Code != testCase.errCode2 {
				t.Errorf("token validation failed: expected %v got %v", testCase.errCode1, appErr.Code)
			}
		}
	}
}

func TestTokenManager_IssueValidateScopedToken(t *testing.T) {
	testCases := []TestCase{
		{
			name:                "issue and validate token everything valid",
			tm:                  createTokenManager(),
			tm2:                 createTokenManager(),
			subject:             testutils.GenerateRandomString(12),
			scope:               token.TokenScopeChangeEmail,
			wantErrAtIssuing:    false,
			wantErrAtValidating: false,
		},
		{
			name: "issue and validate token with invalid jwt secret",
			tm: createTokenManagerWithCustom(&token.Config{
				JwtSecret: []byte("test_secret_invalid"),
			}),
			tm2:                 createTokenManager(),
			subject:             testutils.GenerateRandomString(12),
			scope:               token.TokenScopeChangeEmail,
			wantErrAtIssuing:    false,
			wantErrAtValidating: true,
			errCode2:            apperror.CodeTokenInvalidSignature,
		},
		{
			name: "issue and validate with very short expiry time",
			tm: createTokenManagerWithCustom(&token.Config{
				ScopedTokenTTL: time.Microsecond,
			}),
			tm2:                 createTokenManager(), // NOTE: using another token manager won't lead to any error
			subject:             testutils.GenerateRandomString(12),
			scope:               token.TokenScopeChangeEmail,
			wantErrAtIssuing:    false,
			wantErrAtValidating: true,
			errCode2:            apperror.CodeTokenExpired,
		},
		{
			name:                "issue and validate token with invalid scope",
			tm:                  createTokenManager(),
			tm2:                 createTokenManager(),
			subject:             testutils.GenerateRandomString(12),
			scope:               "invalid",
			wantErrAtIssuing:    true,
			errCode1:            apperror.CodeTokenInvalid,
			wantErrAtValidating: false,
		},
	}
	for _, testCase := range testCases {
		tokenStr, appErr := testCase.tm.IssueScopedToken(context.Background(), testCase.subject, testCase.scope)
		if testCase.wantErrAtIssuing {
			if appErr == nil {
				t.Errorf("token validation failed: expected error got nil")
				os.Exit(1)
			}

			if appErr.Code != testCase.errCode1 {
				t.Errorf("token validation failed: expected %v got %v", testCase.errCode1, appErr.Code)
			}
		}

		_, appErr = testCase.tm2.ValidateScopedToken(context.Background(), tokenStr)
		if testCase.wantErrAtValidating {
			if appErr == nil {
				t.Errorf("token validation failed: expected error got nil")
				os.Exit(1)
			}

			if appErr.Code != testCase.errCode2 {
				t.Errorf("token validation failed: expected %v got %v", testCase.errCode1, appErr.Code)
			}
		}
	}
}

func TestTokenManager_RevokeAuthToken(t *testing.T) {
	testCases := []TestCase{
		{
			name:                "issue and validate token everything valid",
			tm:                  createTokenManager(),
			tm2:                 createTokenManager(),
			subject:             testutils.GenerateRandomString(12),
			sessionID:           testutils.GenerateRandomString(12),
			wantErrAtIssuing:    false,
			wantErrAtValidating: false,
		},
	}

	for _, testCase := range testCases {
		_, appErr := testCase.tm.IssueScopedToken(context.Background(), testCase.subject, testCase.scope)
		if testCase.wantErrAtIssuing {
			if appErr == nil {
				t.Errorf("token validation failed: expected error got nil")
				os.Exit(1)
			}

			if appErr.Code != testCase.errCode1 {
				t.Errorf("token validation failed: expected %v got %v", testCase.errCode1, appErr.Code)
			}
		}

		appErr = testCase.tm2.RevokeAuthToken(context.Background(), testCase.subject, testCase.sessionID)
		if testCase.wantErrAtValidating {
			if appErr == nil {
				t.Errorf("token validation failed: expected error got nil")
				os.Exit(1)
			}
		}
	}
}

func TestTokenManager_RevokeScopedToken(t *testing.T) {
	testCases := []TestCase{
		{
			name:                "issue and validate token everything valid",
			tm:                  createTokenManager(),
			tm2:                 createTokenManager(),
			subject:             testutils.GenerateRandomString(12),
			scope:               token.TokenScopeChangeEmail,
			wantErrAtIssuing:    false,
			wantErrAtValidating: false,
		},
	}

	for _, testCase := range testCases {
		tokenStr, appErr := testCase.tm.IssueScopedToken(context.Background(), testCase.subject, testCase.scope)
		if testCase.wantErrAtIssuing {
			if appErr == nil {
				t.Errorf("token validation failed: expected error got nil")
				os.Exit(1)
			}
			if appErr.Code != testCase.errCode1 {
				t.Errorf("token validation failed: expected %v got %v", testCase.errCode1, appErr.Code)
			}
		}

		appErr = testCase.tm2.RevokeScopedToken(context.Background(), tokenStr)
		if testCase.wantErrAtValidating {
			if appErr == nil {
				t.Errorf("token validation failed: expected error got nil")
				os.Exit(1)
			}
		}
	}
}

func TestTokenManager_RevokeOtherUserTokens(t *testing.T) {
	tm := createTokenManager()
	subject1 := testutils.GenerateRandomString(12)
	sessionID1 := testutils.GenerateRandomString(12)
	sessionID2 := testutils.GenerateRandomString(12)
	sessionID3 := testutils.GenerateRandomString(12)

	tokenStr1, appErr := tm.IssueAuthToken(context.Background(), subject1, sessionID1, token.TokenScopeAuth)
	if appErr != nil {
		t.Errorf("token validation failed: expected error got nil")
	}
	tokenStr2, appErr := tm.IssueAuthToken(context.Background(), subject1, sessionID2, token.TokenScopeAuth)
	if appErr != nil {
		t.Errorf("token validation failed: expected error got nil")
	}
	tokenStr3, appErr := tm.IssueAuthToken(context.Background(), subject1, sessionID3, token.TokenScopeAuth)
	if appErr != nil {
		t.Errorf("token validation failed: expected error got nil")
	}

	// Revoking
	appErr = tm.RevokeOtherUserTokens(context.Background(), subject1, sessionID1)
	if appErr != nil {
		t.Errorf("token validation failed: expected error got nil")
	}

	_, appErr = tm.ValidateAuthToken(context.Background(), tokenStr1)
	if appErr != nil {
		t.Errorf("token validation failed: expected error got nil")
	}
	_, appErr = tm.ValidateAuthToken(context.Background(), tokenStr2)
	if appErr == nil {
		t.Errorf("token validation failed: expected error got nil")
	}
	_, appErr = tm.ValidateAuthToken(context.Background(), tokenStr3)
	if appErr == nil {
		t.Errorf("token validation failed: expected error got nil")
	}
}

func TestTokenManager_RevokeAllUserTokens(t *testing.T) {
	tm := createTokenManager()
	subject1 := testutils.GenerateRandomString(12)
	sessionID1 := testutils.GenerateRandomString(12)
	sessionID2 := testutils.GenerateRandomString(12)
	sessionID3 := testutils.GenerateRandomString(12)

	tokenStr1, appErr := tm.IssueAuthToken(context.Background(), subject1, sessionID1, token.TokenScopeAuth)
	if appErr != nil {
		t.Errorf("token validation failed: expected error got nil")
	}
	tokenStr2, appErr := tm.IssueAuthToken(context.Background(), subject1, sessionID2, token.TokenScopeAuth)
	if appErr != nil {
		t.Errorf("token validation failed: expected error got nil")
	}
	tokenStr3, appErr := tm.IssueAuthToken(context.Background(), subject1, sessionID3, token.TokenScopeAuth)
	if appErr != nil {
		t.Errorf("token validation failed: expected error got nil")
	}

	// Revoking
	appErr = tm.RevokeAllUserTokens(context.Background(), subject1)
	if appErr != nil {
		t.Errorf("token validation failed: expected error got nil")
	}

	_, appErr = tm.ValidateAuthToken(context.Background(), tokenStr1)
	if appErr == nil {
		t.Errorf("token validation failed: expected error got nil")
	}
	_, appErr = tm.ValidateAuthToken(context.Background(), tokenStr2)
	if appErr == nil {
		t.Errorf("token validation failed: expected error got nil")
	}
	_, appErr = tm.ValidateAuthToken(context.Background(), tokenStr3)
	if appErr == nil {
		t.Errorf("token validation failed: expected error got nil")
	}
}

func createTokenManager() *token.TokenManager {
	redisClient := testutils.MustConnectRedis()
	return token.NewTokenManager(token.Config{
		JwtSecret:      []byte("test_secret"),
		SessionTTL:     time.Minute * 15,
		ScopedSecret:   []byte("test_scoped_secret"),
		ScopedTokenTTL: time.Minute * 10,
	}, redisClient)
}

func createTokenManagerWithCustom(cfg *token.Config) *token.TokenManager {
	redisClient := testutils.MustConnectRedis()

	var jwtSecret []byte
	if cfg.JwtSecret != nil {
		jwtSecret = cfg.JwtSecret
	} else {
		jwtSecret = []byte("test_secret")
	}

	var sessionTTL time.Duration
	if cfg.SessionTTL != 0 {
		sessionTTL = cfg.SessionTTL
	} else {
		sessionTTL = time.Minute * 15
	}

	var scopedSecret []byte
	if cfg.ScopedSecret != nil {
		scopedSecret = cfg.ScopedSecret
	} else {
		scopedSecret = []byte("test_scoped_secret")
	}

	var scopedTokenTTL time.Duration
	if cfg.ScopedTokenTTL != 0 {
		scopedTokenTTL = cfg.ScopedTokenTTL
	} else {
		scopedTokenTTL = time.Minute * 10
	}

	return token.NewTokenManager(token.Config{
		JwtSecret:      jwtSecret,
		SessionTTL:     sessionTTL,
		ScopedSecret:   scopedSecret,
		ScopedTokenTTL: scopedTokenTTL,
	}, redisClient)
}
