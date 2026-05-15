package protoutils_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/protoutils"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
)

func TestValidatePaginationReq(t *testing.T) {
	tests := []struct {
		name        string
		req         *corev1.PaginationRequest
		wantCode    apperror.ErrorCode
		wantMessage string
	}{
		{
			name: "valid minimum values",
			req: &corev1.PaginationRequest{
				Page:  1,
				Limit: 1,
			},
		},
		{
			name: "valid maximum limit",
			req: &corev1.PaginationRequest{
				Page:  5,
				Limit: protoutils.MaxPaginationLimit,
			},
		},
		{
			name:        "nil request",
			req:         nil,
			wantCode:    apperror.CodeValidation,
			wantMessage: "pagination request cannot be nil",
		},
		{
			name: "page less than one",
			req: &corev1.PaginationRequest{
				Page:  0,
				Limit: 10,
			},
			wantCode:    apperror.CodeValidation,
			wantMessage: "page number must be greater than 0",
		},
		{
			name: "limit less than one",
			req: &corev1.PaginationRequest{
				Page:  1,
				Limit: 0,
			},
			wantCode:    apperror.CodeValidation,
			wantMessage: "limit must be at least 1",
		},
		{
			name: "limit exceeds maximum",
			req: &corev1.PaginationRequest{
				Page:  1,
				Limit: protoutils.MaxPaginationLimit + 1,
			},
			wantCode:    apperror.CodeValidation,
			wantMessage: "limit cannot exceed 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := protoutils.ValidatePaginationReq(tt.req)

			if tt.wantMessage == "" {
				if err != nil {
					t.Fatalf("ValidatePaginationReq() returned unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("ValidatePaginationReq() returned nil error")
			}

			if err.Code != tt.wantCode {
				t.Fatalf("ValidatePaginationReq() code = %q, want %q", err.Code, tt.wantCode)
			}

			if err.Message != tt.wantMessage {
				t.Fatalf("ValidatePaginationReq() message = %q, want %q", err.Message, tt.wantMessage)
			}
		})
	}
}

func TestValidateIDAndScopedTokenReq(t *testing.T) {
	tests := []struct {
		name        string
		req         *corev1.IDAndScopedTokenRequest
		wantCode    apperror.ErrorCode
		wantMessage string
	}{
		{
			name: "valid request",
			req: &corev1.IDAndScopedTokenRequest{
				Id:         uuid.NewString(),
				TokenScope: string(token.TokenScopeAdmin),
			},
		},
		{
			name:        "nil request",
			req:         nil,
			wantCode:    apperror.CodeValidation,
			wantMessage: "request body cannot be nil",
		},
		{
			name: "invalid uuid",
			req: &corev1.IDAndScopedTokenRequest{
				Id:         "not-a-uuid",
				TokenScope: string(token.TokenScopeAdmin),
			},
			wantCode:    apperror.CodeValidation,
			wantMessage: "provided id is not a valid uuid",
		},
		{
			name: "missing token scope",
			req: &corev1.IDAndScopedTokenRequest{
				Id: uuid.NewString(),
			},
			wantCode:    apperror.CodeValidation,
			wantMessage: "token scope must be provided",
		},
		{
			name: "invalid token scope",
			req: &corev1.IDAndScopedTokenRequest{
				Id:         uuid.NewString(),
				TokenScope: "INVALID_SCOPE",
			},
			wantCode:    apperror.CodeValidation,
			wantMessage: "token scope must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := protoutils.ParseIDAndScopedTokenReq(tt.req)

			if tt.wantMessage == "" {
				if err != nil {
					t.Fatalf("ValidateIDAndScopedTokenReq() returned unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("ValidateIDAndScopedTokenReq() returned nil error")
			}

			if err.Code != tt.wantCode {
				t.Fatalf("ValidateIDAndScopedTokenReq() code = %q, want %q", err.Code, tt.wantCode)
			}

			if err.Message != tt.wantMessage {
				t.Fatalf("ValidateIDAndScopedTokenReq() message = %q, want %q", err.Message, tt.wantMessage)
			}
		})
	}
}
